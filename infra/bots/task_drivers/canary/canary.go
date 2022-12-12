// Copyright 2020 The Chromium Authors. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be
// found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strconv"
	"time"

	"cloud.google.com/go/datastore"

	"go.skia.org/infra/autoroll/go/manual"
	"go.skia.org/infra/go/auth"
	"go.skia.org/infra/go/firestore"
	"go.skia.org/infra/go/gerrit"
	"go.skia.org/infra/go/git"
	"go.skia.org/infra/go/skerr"
	"go.skia.org/infra/task_driver/go/lib/auth_steps"
	"go.skia.org/infra/task_driver/go/lib/checkout"
	"go.skia.org/infra/task_driver/go/td"
)

const (
	MinWaitDuration = 4 * time.Minute
)

var (
	canaryRollNotCreatedErr        = errors.New("Canary roll could not be created. Ask the Infra Gardener to investigate (or directly ping rmistry@) if this happens consistently.")
	canaryRollSuccessTooQuicklyErr = fmt.Errorf("Canary roll returned success in less than %s. Failing canary due to skbug.com/10563.", MinWaitDuration)

	// Lets add the roll link only once to step data.
	addedRollLinkStepData = false
)

func main() {
	var (
		projectId = flag.String("project_id", "", "ID of the Google Cloud project.")
		taskId    = flag.String("task_id", "", "ID of this task.")
		taskName  = flag.String("task_name", "", "Name of the task.")
		output    = flag.String("o", "", "If provided, dump a JSON blob of step data to the given file. Prints to stdout if '-' is given.")
		local     = flag.Bool("local", true, "True if running locally (as opposed to on the bots)")

		checkoutFlags = checkout.SetupFlags(nil)

		rollerName           = flag.String("roller_name", "", "The roller we will use to create the canary with.")
		canaryCQKeyword      = flag.String("cq_keyword", "", "The canary's CQ keyword. Eg: Canary-Chromium-CL.")
		targetProjectBaseURL = flag.String("target_project_base_url", "", "The base URL to append the canaryCQKeyword's value to. canaryCQKeyword must be specified for this to be used. Eg: https://chromium-review.googlesource.com/c/")
	)
	ctx := td.StartRun(projectId, taskId, taskName, output, local)
	defer td.EndRun(ctx)
	if *rollerName == "" {
		td.Fatalf(ctx, "--roller_name must be specified")
	}

	rs, err := checkout.GetRepoState(checkoutFlags)
	if err != nil {
		td.Fatal(ctx, skerr.Wrap(err))
	}
	if rs.Issue == "" || rs.Patchset == "" {
		td.Fatalf(ctx, "This task driver should be run only as a try bot")
	}

	// Create token source with scope for datastore access.
	ts, err := auth_steps.Init(ctx, *local, auth.ScopeUserinfoEmail, datastore.ScopeDatastore)
	if err != nil {
		td.Fatal(ctx, skerr.Wrap(err))
	}

	// Add documentation link for canary rolls.
	td.StepText(ctx, "Canary roll doc", "https://goto.google.com/autoroller-canary-bots")

	// Instantiate Gerrit.
	client, _, err := auth_steps.InitHttpClient(ctx, *local, gerrit.AuthScope)
	if err != nil {
		td.Fatal(ctx, skerr.Wrap(err))
	}
	g, err := gerrit.NewGerrit("https://skia-review.googlesource.com", client)
	if err != nil {
		td.Fatal(ctx, skerr.Wrap(err))
	}

	// Read footers from Gerrit change if canaryCQKeyword is specified.
	canaryCQKeywordValue := ""
	if *canaryCQKeyword != "" {
		issueNum, err := strconv.ParseInt(rs.Issue, 10, 64)
		if err != nil {
			td.Fatal(ctx, skerr.Wrap(err))
		}
		commit, err := g.GetCommit(ctx, issueNum, rs.Patchset)
		if err != nil {
			td.Fatal(ctx, skerr.Wrap(err))
		}
		footersMap := git.GetFootersMap(commit.Message)
		canaryCQKeywordValue = git.GetStringFooterVal(footersMap, *canaryCQKeyword)
	}

	// Instantiate firestore DB.
	manualRollDB, err := manual.NewDBWithParams(ctx, firestore.FIRESTORE_PROJECT, "production", ts)
	if err != nil {
		td.Fatal(ctx, skerr.Wrap(err))
	}

	// Retry if canary roll could not be created or if canary roll returned
	// success too quickly (skbug.com/10563).
	retryAttempts := 3
	for retry := 0; ; retry++ {
		retryText := ""
		if retry > 0 {
			retryText = fmt.Sprintf(" (Retry #%d)", retry)
		}

		req := manual.ManualRollRequest{
			Requester:        *rollerName,
			RollerName:       *rollerName,
			Status:           manual.STATUS_PENDING,
			Timestamp:        firestore.FixTimestamp(time.Now()),
			Revision:         rs.GetPatchRef(),
			ExternalChangeId: canaryCQKeywordValue,

			DryRun:            true,
			NoEmail:           true,
			NoResolveRevision: true,
			Canary:            true,
		}
		if err := td.Do(ctx, td.Props(fmt.Sprintf("Trigger canary roll%s", retryText)).Infra(), func(ctx context.Context) error {
			return manualRollDB.Put(&req)
		}); err != nil {
			// Immediately fail for errors in triggering.
			td.Fatal(ctx, skerr.Wrap(err))
		}

		if err := waitForCanaryRoll(ctx, manualRollDB, req.Id, fmt.Sprintf("Wait for canary roll%s", retryText), canaryCQKeywordValue, *targetProjectBaseURL); err != nil {
			// Retry these errors.
			if err == canaryRollNotCreatedErr || err == canaryRollSuccessTooQuicklyErr {
				if retry >= (retryAttempts - 1) {
					td.Fatal(ctx, skerr.Wrapf(err, "failed inspite of 3 retries"))
				}
				time.Sleep(time.Minute)
				continue
			}
			// Immediately fail for all other errors.
			td.Fatal(ctx, skerr.Wrap(err))
		} else {
			// The canary roll was successful, break out of the
			// retry loop.
			break
		}
	}
}

func waitForCanaryRoll(parentCtx context.Context, manualRollDB manual.DB, rollId, stepName, canaryCQKeywordValue, targetProjectBaseURL string) error {
	ctx := td.StartStep(parentCtx, td.Props(stepName))
	defer td.EndStep(ctx)
	startTime := time.Now()

	// For writing to the step's log stream.
	stdout := td.NewLogStream(ctx, "stdout", td.SeverityInfo)
	for {
		roll, err := manualRollDB.Get(ctx, rollId)
		if err != nil {
			return td.FailStep(ctx, fmt.Errorf("Could not find canary roll with ID: %s", rollId))
		}
		cl := roll.Url
		var rollStatus string
		if cl == "" {
			rollStatus = fmt.Sprintf("Canary roll has status %s", roll.Status)
		} else {
			if !addedRollLinkStepData {
				// Add the roll link to both the current step and it's parent.
				td.StepText(ctx, "Canary roll CL", cl)
				td.StepText(parentCtx, "Canary roll CL", cl)
				if canaryCQKeywordValue != "" && targetProjectBaseURL != "" {
					// Display link to additional patch.
					td.StepText(ctx, "Additional Patch", targetProjectBaseURL+canaryCQKeywordValue)
					td.StepText(parentCtx, "Additional Patch", targetProjectBaseURL+canaryCQKeywordValue)
				}
				addedRollLinkStepData = true
			}
			rollStatus = fmt.Sprintf("Canary roll [ %s ] has status %s", roll.Url, roll.Status)
		}
		if _, err := stdout.Write([]byte(rollStatus)); err != nil {
			return td.FailStep(ctx, fmt.Errorf("Could not write to stdout: %s", err))
		}

		if roll.Status == manual.STATUS_COMPLETE {
			if roll.Result == manual.RESULT_SUCCESS {
				// This is a hopefully temperory workaround for skbug.com/10563. Sometimes
				// Canary-Chromium returns success immediately after creating a change and
				// before the tryjobs have a chance to run. If we have waited
				// for < MinWaitDuration then be cautious and assume failure.
				if time.Now().Before(startTime.Add(MinWaitDuration)) {
					return td.FailStep(ctx, canaryRollSuccessTooQuicklyErr)
				}
				return nil
			} else if roll.Result == manual.RESULT_FAILURE {
				if cl == "" {
					return td.FailStep(ctx, canaryRollNotCreatedErr)
				}
				return td.FailStep(ctx, fmt.Errorf("Canary roll [ %s ] failed", cl))
			} else if roll.Result == manual.RESULT_UNKNOWN {
				return td.FailStep(ctx, fmt.Errorf("Canary roll [ %s ] completed with an unknown result", cl))
			}
		}
		time.Sleep(30 * time.Second)
	}
}
