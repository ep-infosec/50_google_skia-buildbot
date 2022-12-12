// Package cq provides tools for interacting with the CQ tools.
package cq

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/golang/protobuf/ptypes"
	buildbucketpb "go.chromium.org/luci/buildbucket/proto"

	"go.skia.org/infra/go/common"
	"go.skia.org/infra/go/gerrit"
	"go.skia.org/infra/go/git"
	"go.skia.org/infra/go/git/git_common"
	"go.skia.org/infra/go/gitiles"
	"go.skia.org/infra/go/metrics2"
	"go.skia.org/infra/go/skerr"
	"go.skia.org/infra/go/sklog"
	"go.skia.org/infra/go/util"
	"go.skia.org/infra/task_scheduler/go/specs"
)

const (
	MAIN_REF = git_common.RefsHeadsPrefix + git.MainBranch

	// Constants for in-flight metrics.
	INFLIGHT_METRIC_NAME     = "in_flight"
	INFLIGHT_TRYBOT_DURATION = "trybot_duration"
	INFLIGHT_TRYBOT_NUM      = "trybot_num"
	INFLIGHT_WAITING_IN_CQ   = "waiting_in_cq"

	// Constants for landed metrics.
	LANDED_METRIC_NAME     = "after_commit"
	LANDED_TRYBOT_DURATION = "trybot_duration"
	LANDED_TOTAL_DURATION  = "total_duration"
)

var (
	// Slice of all known presubmit bot names.
	PRESUBMIT_BOTS = []string{"skia_presubmit-Trybot"}

	// Mutext to control access to the slice of CQ trybots.
	cqTryBotsMutex sync.RWMutex
)

// NewClient creates a new client for interacting with CQ tools.
func NewClient(gerritClient *gerrit.Gerrit, cqTryBotsFunc GetCQTryBotsFn, metricName string) (*Client, error) {
	cqTryBots, err := cqTryBotsFunc()
	if err != nil {
		return nil, err
	}
	return &Client{gerritClient, util.NewStringSet(cqTryBots), cqTryBotsFunc, metricName}, err
}

// GetCQTryBotsFn is an interface for returing the CQ trybots of a project.
type GetCQTryBotsFn func() ([]string, error)

type Client struct {
	gerritClient  *gerrit.Gerrit
	cqTryBots     util.StringSet
	cqTryBotsFunc GetCQTryBotsFn
	metricName    string
}

// GetSkiaCQTryBots is a Skia implementation of GetCQTryBotsFn.
func GetSkiaCQTryBots() ([]string, error) {
	cfg, err := GetCQJobsToConfig(gitiles.NewRepo(common.REPO_SKIA, nil), MAIN_REF)
	if err != nil {
		return nil, err
	}
	return GetCQTryBots(cfg)
}

// GetSkiaInfraCQTryBots is a Skia Infra implementation of GetCQTryBotsFn.
func GetSkiaInfraCQTryBots() ([]string, error) {
	cfg, err := GetCQJobsToConfig(gitiles.NewRepo(common.REPO_SKIA_INFRA, nil), MAIN_REF)
	if err != nil {
		return nil, err
	}
	return GetCQTryBots(cfg)
}

// GetCQJobsToConfig returns the Config for the given repo.
func GetCQJobsToConfig(repo *gitiles.Repo, ref string) (map[string]*specs.CommitQueueJobConfig, error) {
	contents, err := repo.ReadFileAtRef(context.Background(), specs.TASKS_CFG_FILE, ref)
	if err != nil {
		return nil, err
	}
	tasksCfg, err := specs.ParseTasksCfg(string(contents))
	if err != nil {
		return nil, skerr.Wrapf(err, "Error when parsing tasks.json cfg")
	}
	return tasksCfg.CommitQueue, nil
}

// GetCQTryBots is a convenience method for retrieving the list of CQ trybots
// from a Config.
func GetCQTryBots(cqJobsToConfig map[string]*specs.CommitQueueJobConfig) ([]string, error) {
	tryJobs := []string{}
	for tryJob := range cqJobsToConfig {
		if util.ContainsAny(tryJob, PRESUBMIT_BOTS) {
			// Exclude presubmit bots because they could fail or be delayed
			// due to factors such as owners approval and other project
			// specific checks.
			continue
		}
		tryJobs = append(tryJobs, tryJob)
	}

	sklog.Infof("The list of CQ trybots is: %s", tryJobs)
	return tryJobs, nil
}

// RefreshCQTryBots refreshes the slice of CQ trybots on the instance. Access
// to the trybots is protected by a RWMutex.
func (c *Client) RefreshCQTryBots() error {
	tryBots, err := c.cqTryBotsFunc()
	if err != nil {
		return err
	}
	cqTryBotsMutex.Lock()
	defer cqTryBotsMutex.Unlock()
	c.cqTryBots = util.NewStringSet(tryBots)
	return nil
}

// ReportCQStats reports all relevant stats for the specified Gerrit change.
// Note: Different stats are reported depending on whether the change has been
// merged or not.
// All created metrics will be registered in reportedMetrics.
func (c *Client) ReportCQStats(ctx context.Context, change int64, reportedMetrics map[metrics2.Int64Metric]struct{}) error {
	changeInfo, err := c.gerritClient.GetIssueProperties(ctx, change)
	if err != nil {
		return err
	}
	patchsetIds := changeInfo.GetPatchsetIDs()
	latestPatchsetId := patchsetIds[len(patchsetIds)-1]
	if changeInfo.Committed {
		// TODO(rmistry): The last patchset in Gerrit does not contain trybot
		// information so we have to look at the one immediately before it.
		// This will be fixed with crbug.com/634944.
		latestPatchsetId = patchsetIds[len(patchsetIds)-2]
	}

	builds, err := c.gerritClient.GetTrybotResults(ctx, change, latestPatchsetId)
	if err != nil {
		return err
	}
	// Consider only CQ bots.
	cqBuilds := []*buildbucketpb.Build{}
	for _, b := range builds {
		if c.isCQTryBot(b.Builder.Builder) {
			cqBuilds = append(cqBuilds, b)
		}
	}
	gerritURL := fmt.Sprintf("%s/c/%d/%d", gerrit.GerritSkiaURL, change, latestPatchsetId)
	if len(cqBuilds) == 0 {
		sklog.Infof("No trybot results were found for %s", gerritURL)
		return nil
	}

	sklog.Infof("Starting processing %s. Merged status: %t", gerritURL, changeInfo.Committed)

	if changeInfo.Committed {
		c.ReportCQStatsForLandedCL(cqBuilds, gerritURL, reportedMetrics)
	} else {
		c.ReportCQStatsForInFlightCL(cqBuilds, gerritURL, reportedMetrics)
	}
	return nil
}

// ReportCQStatsForLandedCL reports the following metrics for the specified
// change and patchsetID:
// * The total time the change spent waiting for CQ trybots to complete.
// * The time each CQ trybot took to complete.
// All created metrics will be registered in reportedMetrics.
func (c *Client) ReportCQStatsForLandedCL(cqBuilds []*buildbucketpb.Build, gerritURL string, reportedMetrics map[metrics2.Int64Metric]struct{}) {
	endTimeOfCQBots := time.Time{}
	maximumTrybotDuration := int64(0)
	for _, b := range cqBuilds {
		createdTime, err := ptypes.Timestamp(b.CreateTime)
		if err != nil {
			sklog.Errorf("Failed to convert timestamp for %d; skipping: %s", b.Id, err)
			continue
		}
		createdTime = createdTime.UTC()
		if b.EndTime == nil {
			sklog.Warningf("Skipping %s on %s. The correct completed time has not shown up in Buildbucket yet.", b.Builder.Builder, gerritURL)
			continue
		}
		completedTime, err := ptypes.Timestamp(b.EndTime)
		if err != nil {
			sklog.Errorf("Failed to convert timestamp for %d; skipping: %s", b.Id, err)
			continue
		}
		completedTime = completedTime.UTC()
		if endTimeOfCQBots.Before(completedTime) {
			endTimeOfCQBots = completedTime
		}

		duration := int64(completedTime.Sub(createdTime).Seconds())
		sklog.Infof("%s was created at %s by %s and completed at %s. Total duration: %d", b.Builder.Builder, createdTime, gerritURL, completedTime, duration)
		landedTrybotDurationMetric := c.getLandedTrybotDurationMetric(b.Builder.Builder, gerritURL)
		landedTrybotDurationMetric.Update(duration)
		reportedMetrics[landedTrybotDurationMetric] = struct{}{}

		if duration > maximumTrybotDuration {
			maximumTrybotDuration = duration
		}
	}

	sklog.Infof("Maximum trybot duration for %s: %d", gerritURL, maximumTrybotDuration)
	sklog.Infof("Furthest completion time for %s: %s", gerritURL, endTimeOfCQBots)
	landedTotalDurationMetric := metrics2.GetInt64Metric(fmt.Sprintf("%s_%s_%s", c.metricName, LANDED_METRIC_NAME, LANDED_TOTAL_DURATION), map[string]string{"gerritURL": gerritURL})
	landedTotalDurationMetric.Update(maximumTrybotDuration)
	reportedMetrics[landedTotalDurationMetric] = struct{}{}
}

// ReportCQStatsForInFlightCL reports the following metrics for the specified
// change and patchsetID:
// * How long CQ trybots have been running for.
// * How many CQ trybots have been triggered.
// All created metrics will be registered in reportedMetrics.
func (c *Client) ReportCQStatsForInFlightCL(cqBuilds []*buildbucketpb.Build, gerritURL string, reportedMetrics map[metrics2.Int64Metric]struct{}) {
	totalTriggeredCQBots := int(0)
	currentTime := time.Now()
	for _, b := range cqBuilds {
		totalTriggeredCQBots++

		createdTime, err := ptypes.Timestamp(b.CreateTime)
		if err != nil {
			sklog.Errorf("Failed to convert timestamp for %d; skipping: %s", b.Id, err)
			continue
		}
		createdTime = createdTime.UTC()
		if b.EndTime != nil {
			if time.Hour*24 < time.Now().UTC().Sub(createdTime) {
				// The build was created more than a day ago. Do not include it
				// in totalTriggeredCQBots. See skbug.com/7340.
				// Creation time is used above instead of completion time because
				// that is what CQ does:
				// https://chrome-internal.googlesource.com/infra/infra_internal/+show/master/infra_internal/services/cq/verification/tryjob_utils.py#1271
				totalTriggeredCQBots--
			}
			// The build has completed so move on.
			continue
		}

		duration := int64(currentTime.Sub(createdTime).Seconds())
		inflightTrybotDurationMetric := c.getInflightTrybotDurationMetric(b.Builder.Builder, gerritURL)
		inflightTrybotDurationMetric.Update(duration)
		reportedMetrics[inflightTrybotDurationMetric] = struct{}{}
	}

	cqTryBotsMutex.RLock()
	cqTryBotsMutex.RUnlock()
	trybotNumDurationMetric := metrics2.GetInt64Metric(fmt.Sprintf("%s_%s_%s", c.metricName, INFLIGHT_METRIC_NAME, INFLIGHT_TRYBOT_NUM), map[string]string{"gerritURL": gerritURL})
	trybotNumDurationMetric.Update(int64(totalTriggeredCQBots))
	reportedMetrics[trybotNumDurationMetric] = struct{}{}
}

func (c *Client) getInflightTrybotDurationMetric(tryBot, gerritURL string) metrics2.Int64Metric {
	metricName := fmt.Sprintf("%s_%s_%s", c.metricName, INFLIGHT_METRIC_NAME, INFLIGHT_TRYBOT_DURATION)
	tags := map[string]string{
		"trybot":    tryBot,
		"gerritURL": gerritURL,
	}
	return metrics2.GetInt64Metric(metricName, tags)
}

func (c *Client) getLandedTrybotDurationMetric(tryBot, gerritURL string) metrics2.Int64Metric {
	metricName := fmt.Sprintf("%s_%s_%s", c.metricName, LANDED_METRIC_NAME, LANDED_TRYBOT_DURATION)
	tags := map[string]string{
		"trybot":    tryBot,
		"gerritURL": gerritURL,
	}
	return metrics2.GetInt64Metric(metricName, tags)
}

func (c *Client) isCQTryBot(builderName string) bool {
	cqTryBotsMutex.RLock()
	isCQTrybot := c.cqTryBots[builderName]
	cqTryBotsMutex.RUnlock()
	return isCQTrybot
}
