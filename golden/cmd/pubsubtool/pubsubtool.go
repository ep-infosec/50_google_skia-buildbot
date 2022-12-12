// The pubsubtool executable is a convenient way to create PubSub topics and subscriptions.
// It also allows for manual injection of messages to test systems end-to-end.
package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/pubsub"
	"cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	"go.skia.org/infra/go/skerr"
	"go.skia.org/infra/go/sklog"
	"go.skia.org/infra/go/sklog/sklogimpl"
	"go.skia.org/infra/go/sklog/stdlogging"
)

func main() {
	bucketName := flag.String("bucket_name", "", "The GCS bucket to listen to (see bucket_notifications)")
	prefix := flag.String("prefix", "", "The GCS prefix to listen to.")
	projectID := flag.String("project_id", "skia-public", "The project for PubSub events")
	topicName := flag.String("topic_name", "", "The topic to create if it does not exist")
	subscriptionName := flag.String("subscription_name", "", "The subscription to create if it does not exist")
	jsonMessageFile := flag.String("json_message_file", "", "A file that contains the JSON contents to send as the body of a pubsub message.")

	sklogimpl.SetLogger(stdlogging.New(os.Stderr))
	flag.Parse()
	task := strings.ToLower(flag.Arg(0))

	ctx := context.Background()
	ts, err := google.DefaultTokenSource(ctx)
	if err != nil {
		exitWithError("getting auth token: %s", err)
	}
	psc, err := pubsub.NewClient(ctx, *projectID, option.WithTokenSource(ts))
	if err != nil {
		exitWithError("Initializing pubsub client for project %s: %s", *projectID, err)
	}

	gsc, err := storage.NewClient(ctx, option.WithTokenSource(ts))
	if err != nil {
		exitWithError("Initializing GCS Client: %s", err)
	}

	if task == "create" {
		if err := createTopicAndSubscription(ctx, psc, *topicName, *subscriptionName); err != nil {
			exitWithError("Making topic %s and subscription %s: %s", *topicName, *subscriptionName, err)
		}
	} else if task == "publish" {
		if err := publishMessage(ctx, psc, *topicName, *jsonMessageFile); err != nil {
			exitWithError("Sending contents of %s to topic %s: %S", *jsonMessageFile)
		}
	} else if task == "bucket_notifications" {
		if err := listBucketNotifications(ctx, gsc, *bucketName); err != nil {
			exitWithError("Listing bucket notifications on GCS bucket %q: %s", *bucketName, err)
		}
	} else if task == "subscribe_to_bucket" {
		if err := subscribeToBucket(ctx, psc, *projectID, *topicName, *subscriptionName, gsc, *bucketName, *prefix); err != nil {
			exitWithError("Creating new bucket notification: %s", err)
		}
	} else {
		exitWithError(`Invalid command: %q. Try "create".`, task)
	}
}

func exitWithError(msg string, args ...interface{}) {
	msg = strings.TrimSuffix(msg, "\n") + "\n"
	fmt.Printf(msg, args...)
	os.Exit(1)
}

func publishMessage(ctx context.Context, psc *pubsub.Client, topic, jsonMessageFile string) error {
	if topic == "" || jsonMessageFile == "" {
		return skerr.Fmt("Can't have empty topic or message file")
	}
	body, err := ioutil.ReadFile(jsonMessageFile)
	if err != nil {
		return skerr.Wrapf(err, "reading %s", jsonMessageFile)
	}
	pr := psc.Topic(topic).Publish(ctx, &pubsub.Message{
		Data: body,
	})
	// Blocks until message actual sent
	_, err = pr.Get(ctx)
	if err != nil {
		return skerr.Wrap(err)
	}
	sklog.Infof("Sent")
	return nil
}

func createTopicAndSubscription(ctx context.Context, psc *pubsub.Client, topic, sub string) error {
	if topic == "" || sub == "" {
		return skerr.Fmt("Can't have empty topic or subscription")
	}
	// Create the topic if it doesn't exist yet.
	t, err := createTopicIfNotExists(ctx, psc, topic)
	if err != nil {
		return skerr.Wrap(err)
	}

	// Create the subscription if it doesn't exist.
	s := psc.Subscription(sub)
	if exists, err := s.Exists(ctx); err != nil {
		return skerr.Wrapf(err, "checking existence of pubsub subscription %q", sub)
	} else if !exists {
		_, err = psc.CreateSubscription(ctx, sub, pubsub.SubscriptionConfig{
			Topic: t,
			// These are the default values for the diff-metrics subscription
			AckDeadline:       2 * time.Minute,
			RetentionDuration: 4 * time.Hour,
			RetryPolicy: &pubsub.RetryPolicy{
				MinimumBackoff: time.Minute,
				MaximumBackoff: 5 * time.Minute,
			},
		})
		if err != nil {
			return skerr.Wrapf(err, "creating pubsub subscription %q", sub)
		}
	}
	sklog.Infof("Topic %s and Subscription %s exist if they didn't before", topic, sub)
	return nil
}

func createTopicIfNotExists(ctx context.Context, psc *pubsub.Client, topic string) (*pubsub.Topic, error) {
	t := psc.Topic(topic)
	if exists, err := t.Exists(ctx); err != nil {
		return nil, skerr.Wrapf(err, "checking whether topic %q exists", topic)
	} else if !exists {
		if t, err = psc.CreateTopic(ctx, topic); err != nil {
			return nil, skerr.Wrapf(err, "creating pubsub topic %q", topic)
		}
	}
	return t, nil
}

func listBucketNotifications(ctx context.Context, gsc *storage.Client, bucketName string) error {
	if bucketName == "" {
		return skerr.Fmt("Must specify bucket")
	}
	bucket := gsc.Bucket(bucketName)
	notifications, err := bucket.Notifications(ctx)
	if err != nil {
		return skerr.Wrap(err)
	}
	sklog.Infof("Retrieved: %d notifications", len(notifications))
	for _, n := range notifications {
		sklog.Infof("%s events under //%s are published to topic %s in project %s", n.EventTypes, n.ObjectNamePrefix, n.TopicID, n.TopicProjectID)
	}
	return nil
}

func subscribeToBucket(ctx context.Context, psc *pubsub.Client, project, topic, subscription string, gsc *storage.Client, bucket, prefix string) error {
	if bucket == "" || prefix == "" {
		return skerr.Fmt("Must specify bucket and prefix")
	}
	t, err := createTopicIfNotExists(ctx, psc, topic)
	if err != nil {
		return skerr.Wrap(err)
	}

	_, err = gsc.Bucket(bucket).AddNotification(ctx, &storage.Notification{
		TopicID:          topic,
		TopicProjectID:   project,
		EventTypes:       []string{storage.ObjectFinalizeEvent},
		ObjectNamePrefix: prefix,
		PayloadFormat:    storage.NoPayload, // We only care about properties
	})
	if err != nil {
		return skerr.Wrapf(err, "creating topic %s in project %s for files from gcs://%s/%s", topic, project, bucket, prefix)
	}

	// Create the subscription if it doesn't exist.
	s := psc.Subscription(subscription)
	if exists, err := s.Exists(ctx); err != nil {
		return skerr.Wrapf(err, "checking existence of pubsub subscription %q", subscription)
	} else if !exists {
		_, err = psc.CreateSubscription(ctx, subscription, pubsub.SubscriptionConfig{
			Topic: t,
			// These are the default values for the data files subscriptions.
			AckDeadline:       2 * time.Minute,
			RetentionDuration: 2 * 24 * time.Hour,
			RetryPolicy: &pubsub.RetryPolicy{
				MinimumBackoff: 10 * time.Second,
				MaximumBackoff: 5 * time.Minute,
			},
			// A deadletter policy should be set up and verified via the
			// cloud console UI (there's usually one additional permission to grant)
			// Retry attempts 5 is usually fine (to prevent bad files from filling up
			// our PubSub queue).
		})
		if err != nil {
			return skerr.Wrapf(err, "creating pubsub subscription %q", subscription)
		}
	} else {
		sklog.Infof("Subscription %q already existed", subscription)
	}
	sklog.Infof("Subscription %s ready to listen to topic %s which gets events from files created in gs://%s/%s", subscription, topic, bucket, prefix)
	return nil
}
