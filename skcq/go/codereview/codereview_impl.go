package codereview

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cenkalti/backoff"

	"go.skia.org/infra/go/gerrit"
	"go.skia.org/infra/go/skerr"
	"go.skia.org/infra/go/sklog"
	"go.skia.org/infra/go/util"
)

const (
	// The tag that SkCQ will apply to all comments published by it.
	AutogeneratedCommentTag = "autogenerated:skcq"

	// The number of changes to search in Gerrit for CQ/dry-runs.
	GerritOpenChangesNum = 100
)

// gerritCodeReview implements the CodeReview interface.
type gerritCodeReview struct {
	gerritClient gerrit.GerritInterface
	cfg          *gerrit.Config
}

// NewGerrit returns a gerritCodeReview instance.
func NewGerrit(httpClient *http.Client, cfg *gerrit.Config, gerritURL string) (CodeReview, error) {
	g, err := gerrit.NewGerritWithConfig(cfg, gerritURL, httpClient)
	if err != nil {
		return nil, err
	}
	return &gerritCodeReview{
		gerritClient: g,
		cfg:          cfg,
	}, nil
}

// AddComment implements the CodeReview interface.
func (gc *gerritCodeReview) AddComment(ctx context.Context, ci *gerrit.ChangeInfo, comment string, notify NotifyOption, notifyReason string) error {
	// Convert SkCQ's NotifyOption to Gerrit's NotifyOption.
	var notifyOption gerrit.NotifyOption
	var notifyDetails gerrit.NotifyDetails
	var attentionSets []*gerrit.AttentionSetInput

	switch notify {
	case NotifyNone:
		notifyOption = gerrit.NotifyNone
	case NotifyOwnerTriggerers:
		// Notify the CQ voters + CL owner.
		notifyOption = gerrit.NotifyOwner
		notifyAccounts := gc.GetCQVoters(ctx, ci)
		if !util.In(ci.Owner.Email, notifyAccounts) {
			notifyAccounts = append(notifyAccounts, ci.Owner.Email)
		}
		notifyDetails = gerrit.NotifyDetails{
			gerrit.RecipientTo: &gerrit.NotifyInfo{
				Accounts: notifyAccounts,
			},
		}
		// Create the attention set.
		for _, na := range notifyAccounts {
			attentionSets = append(attentionSets, &gerrit.AttentionSetInput{User: na, Reason: notifyReason})
		}
	case NotifyOwnerReviewersTriggerers:
		// Notify the CQ voters + CL owner + CL reviewers.
		notifyOption = gerrit.NotifyOwnerReviewers
		notifyAccounts := gc.GetCQVoters(ctx, ci)
		if !util.In(ci.Owner.Email, notifyAccounts) {
			notifyAccounts = append(notifyAccounts, ci.Owner.Email)
		}
		for _, r := range ci.Reviewers.Reviewer {
			notifyAccounts = append(notifyAccounts, r.Email)
		}
		notifyDetails = gerrit.NotifyDetails{
			gerrit.RecipientTo: &gerrit.NotifyInfo{
				Accounts: notifyAccounts,
			},
		}
	}

	// SetReview sometimes returns 404s from Gerrit and retries seems to
	// resolve it. Maybe happens when a Gerrit frontend index is stale.
	//
	// Example flow with the used backoff values and assuming we go over
	// MaxElapsedTime on the 4th try:
	//  request#     retry_interval     randomized_interval
	//  1             1                  [0.5,   1.5]
	//  2             1.5                [0.75,  2.25]
	//  3             2.25               [1.125, 3.375]
	//  4             3.375              backoff.Stop
	exp := &backoff.ExponentialBackOff{
		InitialInterval:     time.Second,
		RandomizationFactor: 0.5,
		Multiplier:          1.5,
		MaxInterval:         2 * time.Second,
		MaxElapsedTime:      5 * time.Second,
		Clock:               backoff.SystemClock,
	}
	addCommentFunc := func() error {
		return gc.gerritClient.SetReview(ctx, ci, comment, map[string]int{}, []string{}, notifyOption, notifyDetails, AutogeneratedCommentTag, 0, attentionSets)
	}
	return skerr.Wrap(backoff.Retry(addCommentFunc, exp))
}

// GetChangeRef implements the CodeReview interface.
func (gc *gerritCodeReview) GetChangeRef(ci *gerrit.ChangeInfo) string {
	return fmt.Sprintf("%s%02d/%d/%d", gerrit.ChangeRefPrefix, ci.Issue%100, ci.Issue, gc.GetLatestPatchSetID(ci))
}

// GetCommitAuthor implements the CodeReview interface.
func (gc *gerritCodeReview) GetCommitAuthor(ctx context.Context, issue int64, revision string) (string, error) {
	commitInfo, err := gc.gerritClient.GetCommit(ctx, issue, revision)
	if err != nil {
		return "", err
	}
	if commitInfo.Author == nil {
		return "", errors.New("commitInfo.Author was nil")
	}
	return commitInfo.Author.Email, nil
}

// GetCommitMessage implements the CodeReview interface.
func (gc *gerritCodeReview) GetCommitMessage(ctx context.Context, issue int64) (string, error) {
	commitInfo, err := gc.gerritClient.GetCommit(ctx, issue, "current")
	if err != nil {
		return "", err
	}
	return commitInfo.Message, nil
}

// GetEarliestEquivalentPatchSetID implements the CodeReview interface.
func (gc *gerritCodeReview) GetEarliestEquivalentPatchSetID(ci *gerrit.ChangeInfo) int64 {
	nonTrivial := ci.GetNonTrivialPatchSets()
	return nonTrivial[len(nonTrivial)-1].Number
}

// GetEquivalentPatchSetIDs implements the CodeReview interface.
func (gc *gerritCodeReview) GetEquivalentPatchSetIDs(ci *gerrit.ChangeInfo, patchsetID int64) []int64 {
	ret := []int64{}
	startChecking := false
	for i := len(ci.Patchsets) - 1; i >= 0; i-- {
		patchSet := ci.Patchsets[i]
		if patchSet.Number == patchsetID {
			startChecking = true
		}
		if startChecking {
			// Keep adding till we reach a CODE_CHANGE and then break out.
			ret = append(ret, patchSet.Number)
			if !util.In(patchSet.Kind, gerrit.TrivialPatchSetKinds) {
				break
			}
		}
	}
	return ret
}

// GetFileNames implements the CodeReview interface.
func (gc *gerritCodeReview) GetFileNames(ctx context.Context, ci *gerrit.ChangeInfo) ([]string, error) {
	return gc.gerritClient.GetFileNames(ctx, ci.Issue, strconv.FormatInt(gc.GetLatestPatchSetID(ci), 10))
}

// GetIssueProperties implements the CodeReview interface.
func (gc *gerritCodeReview) GetIssueProperties(ctx context.Context, issue int64) (*gerrit.ChangeInfo, error) {
	return gc.gerritClient.GetIssueProperties(ctx, issue)
}

// GetLatestPatchSetID implements the CodeReview interface.
func (gc *gerritCodeReview) GetLatestPatchSetID(ci *gerrit.ChangeInfo) int64 {
	patchsetIDs := ci.GetPatchsetIDs()
	return patchsetIDs[len(patchsetIDs)-1]
}

// GetSubmittedTogether implements the CodeReview interface.
func (gc *gerritCodeReview) GetSubmittedTogether(ctx context.Context, ci *gerrit.ChangeInfo) ([]*gerrit.ChangeInfo, error) {
	changes, nonVisibleChanges, err := gc.gerritClient.SubmittedTogether(ctx, ci)
	if err != nil {
		return nil, skerr.Wrapf(err, "Could not get the list of submitted together changes")
	}
	if nonVisibleChanges > 0 {
		return nil, skerr.Fmt("The SkCQ service account does not have access to view some submitted together changes of %d", ci.Issue)
	}
	// Filter out the specified ChangeInfo and return fully filled-in objects.
	fullFilteredChanges := []*gerrit.ChangeInfo{}
	for _, c := range changes {
		if c.Id != ci.Id {
			fullCI, err := gc.gerritClient.GetIssueProperties(ctx, c.Issue)
			if err != nil {
				return nil, skerr.Fmt("Could not get full issue properties of %d", c.Issue)
			}
			fullFilteredChanges = append(fullFilteredChanges, fullCI)
		}
	}
	return fullFilteredChanges, nil
}

// IsCQ implements the CodeReview interface.
func (gc *gerritCodeReview) IsCQ(ctx context.Context, ci *gerrit.ChangeInfo) bool {
	return gc.cfg.CqRunning(ci)
}

// IsDryRun implements the CodeReview interface.
func (gc *gerritCodeReview) IsDryRun(ctx context.Context, ci *gerrit.ChangeInfo) bool {
	return gc.cfg.DryRunRunning(ci)
}

// RemoveFromCQ implements the CodeReview interface.
func (gc *gerritCodeReview) RemoveFromCQ(ctx context.Context, ci *gerrit.ChangeInfo, comment string, notifyReason string) {
	// Delete all CQ+1/CQ+2 votes.
	le := ci.Labels[gerrit.LabelCommitQueue]
	for _, labelDetail := range le.All {
		if labelDetail.Value > 0 {
			if err := gc.gerritClient.DeleteVote(ctx, ci.Issue, gerrit.LabelCommitQueue, labelDetail.AccountID, gerrit.NotifyNone, true); err != nil {
				sklog.Errorf("[%d] Could not remove from CQ: %s", ci.Issue, err)
				return
			}
		}
	}
	// Update the change with a comment.
	if err := gc.AddComment(ctx, ci, comment, NotifyOwnerTriggerers, notifyReason); err != nil {
		sklog.Errorf("[%d] Could not add a comment \"%s\": %s", ci.Issue, comment, err)
		return
	}
}

// Search implements the CodeReview interface.
func (gc *gerritCodeReview) Search(ctx context.Context) ([]*gerrit.ChangeInfo, error) {
	// Searching for open issues term will apply for both CQ and dry-run
	// searches.
	openSearchTerm := gerrit.SearchStatus(gerrit.ChangeStatusOpen)

	// Search for CQ runs.
	searchTermsCQ := []*gerrit.SearchTerm{openSearchTerm}
	for label, val := range gc.gerritClient.Config().SetCqLabels {
		searchTermsCQ = append(searchTermsCQ, gerrit.SearchLabel(label, strconv.Itoa(val)))
	}
	changesCQ, err := gc.gerritClient.Search(ctx, GerritOpenChangesNum, true, searchTermsCQ...)
	if err != nil {
		return nil, skerr.Wrapf(err, "Could not search for CQ issues")
	}

	// Search for dry-runs.
	searchTermsDryRun := []*gerrit.SearchTerm{openSearchTerm}
	for label, val := range gc.gerritClient.Config().SetDryRunLabels {
		searchTermsDryRun = append(searchTermsDryRun, gerrit.SearchLabel(label, strconv.Itoa(val)))
	}
	changesDryRun, err := gc.gerritClient.Search(ctx, GerritOpenChangesNum, true, searchTermsDryRun...)
	if err != nil {
		return nil, skerr.Wrapf(err, "Could not search for dry-run issues")
	}

	// Append them to a single slice.
	matchingChanges := append(changesCQ, changesDryRun...)

	// Loop through the matching changes to make sure we remove duplications.
	alreadySeen := map[int64]bool{}
	filteredChanges := []*gerrit.ChangeInfo{}
	for _, ci := range matchingChanges {
		if _, ok := alreadySeen[ci.Issue]; ok {
			continue
		}
		alreadySeen[ci.Issue] = true
		filteredChanges = append(filteredChanges, ci)
	}
	return filteredChanges, nil
}

// SetReadyForReview implements the CodeReview interface.
func (gc *gerritCodeReview) SetReadyForReview(ctx context.Context, ci *gerrit.ChangeInfo) error {
	return gc.gerritClient.SetReadyForReview(ctx, ci)
}

// Submit implements the CodeReview interface.
func (gc *gerritCodeReview) Submit(ctx context.Context, ci *gerrit.ChangeInfo) error {
	return gc.gerritClient.Submit(ctx, ci)
}

// Url implements the CodeReview interface.
func (gc *gerritCodeReview) Url(issueID int64) string {
	return gc.gerritClient.Url(issueID)
}

// GetRepoUrl implements the CodeReview interface.
func (gc *gerritCodeReview) GetRepoUrl(ci *gerrit.ChangeInfo) string {
	return gc.gerritClient.GetRepoUrl() + "/" + ci.Project
}

// GetCQVoters implements the CodeReview interface.
func (gc *gerritCodeReview) GetCQVoters(ctx context.Context, ci *gerrit.ChangeInfo) []string {
	// Find which CQ label value we are looking for.
	labelValue := gerrit.LabelCommitQueueDryRun
	if gc.IsCQ(ctx, ci) {
		labelValue = gerrit.LabelCommitQueueSubmit
	}

	// Fing all voters of the CQ label value.
	voters := []string{}
	if val, ok := ci.Labels[gerrit.LabelCommitQueue]; ok {
		for _, ld := range val.All {
			if ld.Value == labelValue {
				voters = append(voters, ld.Email)
			}
		}
	}
	return voters
}
