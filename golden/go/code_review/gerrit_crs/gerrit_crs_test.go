package gerrit_crs

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.skia.org/infra/go/gerrit"
	"go.skia.org/infra/go/gerrit/mocks"
	"go.skia.org/infra/go/skerr"
	"go.skia.org/infra/go/testutils"
	"go.skia.org/infra/go/vcsinfo"
	"go.skia.org/infra/golden/go/code_review"
)

func TestGetChangelistSunnyDay(t *testing.T) {

	mgi := &mocks.GerritInterface{}
	defer mgi.AssertExpectations(t)

	const id = "235460"
	ts := time.Date(2019, time.August, 21, 16, 44, 26, 0, time.UTC)
	gci := getOpenChangeInfo()
	mgi.On("GetIssueProperties", testutils.AnyContext, int64(235460)).Return(&gci, nil)

	c := New(mgi)

	cl, err := c.GetChangelist(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, code_review.Changelist{
		SystemID: id,
		Owner:    "test@example.com",
		Status:   code_review.Open,
		Subject:  "[gold] Add more tryjob processing tests",
		Updated:  ts,
	}, cl)
}

func TestGetChangelistLanded(t *testing.T) {

	mgi := &mocks.GerritInterface{}
	defer mgi.AssertExpectations(t)

	const id = "235460"
	ts := time.Date(2019, time.August, 21, 16, 44, 26, 0, time.UTC)
	gci := getOpenChangeInfo()
	gci.Status = gerrit.ChangeStatusMerged
	mgi.On("GetIssueProperties", testutils.AnyContext, int64(235460)).Return(&gci, nil)

	c := New(mgi)

	cl, err := c.GetChangelist(context.Background(), id)
	require.NoError(t, err)
	assert.Equal(t, code_review.Changelist{
		SystemID: id,
		Owner:    "test@example.com",
		Status:   code_review.Landed,
		Subject:  "[gold] Add more tryjob processing tests",
		Updated:  ts,
	}, cl)
}

func TestGetChangelistDoesNotExist(t *testing.T) {

	mgi := &mocks.GerritInterface{}
	defer mgi.AssertExpectations(t)

	const id = "235460"
	mgi.On("GetIssueProperties", testutils.AnyContext, int64(235460)).Return(nil, gerrit.ErrNotFound)

	c := New(mgi)

	_, err := c.GetChangelist(context.Background(), id)
	require.Error(t, err)
	require.Equal(t, code_review.ErrNotFound, err)
}

func TestGetChangelistInvalidID(t *testing.T) {

	mgi := &mocks.GerritInterface{}
	defer mgi.AssertExpectations(t)

	const id = "not_an_integer"
	c := New(mgi)

	_, err := c.GetChangelist(context.Background(), id)
	require.Error(t, err)
	require.Equal(t, invalidID, err)
}

func TestGetChangelistOtherErr(t *testing.T) {

	mgi := &mocks.GerritInterface{}
	defer mgi.AssertExpectations(t)

	const id = "235460"
	mgi.On("GetIssueProperties", testutils.AnyContext, int64(235460)).Return(nil, errors.New("oops, sentient AI"))

	c := New(mgi)

	_, err := c.GetChangelist(context.Background(), id)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fetching CL")
	assert.Contains(t, err.Error(), "oops")
}

const omitPS = ""
const omitOrder = 0

func TestGetPatchset_PatchsetExists_Success(t *testing.T) {

	mgi := &mocks.GerritInterface{}
	defer mgi.AssertExpectations(t)

	gci := getOpenChangeInfo()
	mgi.On("GetIssueProperties", testutils.AnyContext, int64(235460)).Return(&gci, nil)

	c := New(mgi)
	const clID = "235460"
	const psOneID = "993b807277763b351e72d01e6d65461c4bf57981"
	const psFourID = "337da6ea3a14fd2899b39d0a60c6828971c0d883"

	expectedFirstPS := code_review.Patchset{
		SystemID:     psOneID,
		ChangelistID: clID,
		Order:        1,
		GitHash:      psOneID,
		Created:      time.Date(2019, time.August, 21, 14, 26, 43, 0, time.UTC),
	}
	expectedFifthPS := code_review.Patchset{
		SystemID:     psFourID,
		ChangelistID: clID,
		Order:        4,
		GitHash:      psFourID,
		Created:      time.Date(2019, time.August, 21, 16, 28, 38, 0, time.UTC),
	}

	ps, err := c.GetPatchset(context.Background(), clID, psOneID, omitOrder)
	require.NoError(t, err)
	assert.Equal(t, expectedFirstPS, ps)
	ps, err = c.GetPatchset(context.Background(), clID, omitPS, 1)
	require.NoError(t, err)
	assert.Equal(t, expectedFirstPS, ps)

	ps, err = c.GetPatchset(context.Background(), clID, psFourID, omitOrder)
	require.NoError(t, err)
	assert.Equal(t, expectedFifthPS, ps)
	ps, err = c.GetPatchset(context.Background(), clID, omitPS, 4)
	require.NoError(t, err)
	assert.Equal(t, expectedFifthPS, ps)
}

func TestGetPatchset_PatchsetDoesNotExist_ReturnsNotFound(t *testing.T) {

	mgi := &mocks.GerritInterface{}
	defer mgi.AssertExpectations(t)

	gci := getOpenChangeInfo()
	mgi.On("GetIssueProperties", testutils.AnyContext, int64(235460)).Return(&gci, nil)

	c := New(mgi)

	const clID = "235460"
	_, err := c.GetPatchset(context.Background(), clID, "does not exist", omitOrder)
	require.Error(t, err)
	assert.Equal(t, code_review.ErrNotFound, err)
	_, err = c.GetPatchset(context.Background(), clID, omitPS, 1000)
	require.Error(t, err)
	assert.Equal(t, code_review.ErrNotFound, err)
}

func TestGetPatchset_ChangelistDoesNotExist_ReturnsNotFound(t *testing.T) {

	mgi := &mocks.GerritInterface{}
	defer mgi.AssertExpectations(t)

	mgi.On("GetIssueProperties", testutils.AnyContext, int64(235460)).Return(nil, gerrit.ErrNotFound)

	c := New(mgi)

	const clID = "235460"
	_, err := c.GetPatchset(context.Background(), clID, "nope", omitOrder)
	require.Error(t, err)
	assert.Equal(t, code_review.ErrNotFound, err)
	_, err = c.GetPatchset(context.Background(), clID, omitPS, 1)
	require.Error(t, err)
	assert.Equal(t, code_review.ErrNotFound, err)
}

func TestGetPatchset_InvalidIDForChangelist_ReturnsError(t *testing.T) {

	mgi := &mocks.GerritInterface{}
	defer mgi.AssertExpectations(t)

	c := New(mgi)

	const clID = "not_an_integer"
	_, err := c.GetPatchset(context.Background(), clID, "nope", omitOrder)
	require.Error(t, err)
	require.Equal(t, invalidID, err)
}

func TestGetPatchset_OtherError_ReturnsError(t *testing.T) {

	mgi := &mocks.GerritInterface{}
	defer mgi.AssertExpectations(t)

	const id = "235460"
	mgi.On("GetIssueProperties", testutils.AnyContext, int64(235460)).Return(nil, errors.New("oops, sentient AI"))

	c := New(mgi)

	_, err := c.GetPatchset(context.Background(), id, "whatever", 7)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "fetching CL")
	assert.Contains(t, err.Error(), "oops")
}

func TestGetChangelistForCommitSunnyDay(t *testing.T) {

	mgi := &mocks.GerritInterface{}
	defer mgi.AssertExpectations(t)

	const id = "235460"
	const clBody = `blah blah blah
    Reviewed-on: https://chromium-review.googlesource.com/c/chromium/src/+/235460
blah blah blah
`

	mgi.On("ExtractIssueFromCommit", clBody).Return(int64(235460), nil)

	c := New(mgi)

	clID, err := c.GetChangelistIDForCommit(context.Background(), &vcsinfo.LongCommit{
		// This is the only field the implementation cares about.
		Body: clBody,
	})
	require.NoError(t, err)
	assert.Equal(t, id, clID)
}

func TestGetChangelistForCommitBadBody(t *testing.T) {

	mgi := &mocks.GerritInterface{}
	defer mgi.AssertExpectations(t)

	const clBody = `malformed body`

	mgi.On("ExtractIssueFromCommit", clBody).Return(int64(0), skerr.Fmt("nope"))

	c := New(mgi)

	_, err := c.GetChangelistIDForCommit(context.Background(), &vcsinfo.LongCommit{
		// This is the only field the implementation cares about.
		Body: clBody,
	})
	require.Error(t, err)
	assert.Equal(t, err, code_review.ErrNotFound)
}

func TestCommentOnChangelistSunnyDay(t *testing.T) {

	mgi := &mocks.GerritInterface{}
	defer mgi.AssertExpectations(t)

	const id = "235460"
	gci := getOpenChangeInfo()
	mgi.On("GetIssueProperties", testutils.AnyContext, int64(235460)).Return(&gci, nil)
	mgi.On("AddComment", testutils.AnyContext, &gci, "blurb").Return(nil)

	c := New(mgi)

	err := c.CommentOn(context.Background(), id, "blurb")
	require.NoError(t, err)
}

func TestCommentOnChangelistGerritError(t *testing.T) {

	mgi := &mocks.GerritInterface{}
	defer mgi.AssertExpectations(t)

	const id = "235460"
	gci := getOpenChangeInfo()
	mgi.On("GetIssueProperties", testutils.AnyContext, int64(235460)).Return(&gci, nil)
	mgi.On("AddComment", testutils.AnyContext, &gci, "blurb").Return(errors.New("internet broke"))

	c := New(mgi)

	err := c.CommentOn(context.Background(), id, "blurb")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "internet broke")
}

func TestCommentOnChangelist_CLNotFound_ReturnsError(t *testing.T) {

	mgi := &mocks.GerritInterface{}

	const id = "235460"
	// This error is based on a real error, which might not be gerrit.ErrNotFound for reasons
	// that are unclear.
	mgi.On("GetIssueProperties", testutils.AnyContext, int64(235460)).Return(nil, errors.New("Got status 404 Not Found (404);"))

	c := New(mgi)

	err := c.CommentOn(context.Background(), id, "blurb")
	require.Error(t, err)
	assert.Equal(t, code_review.ErrNotFound, err)
}

func TestCommentOnChangelist_CLNotFound2_ReturnsError(t *testing.T) {

	mgi := &mocks.GerritInterface{}

	const id = "235460"
	gci := getOpenChangeInfo()
	mgi.On("GetIssueProperties", testutils.AnyContext, int64(235460)).Return(&gci, nil)
	mgi.On("AddComment", testutils.AnyContext, &gci, "blurb").Return(errors.New("Got status 404 Not Found (404);"))

	c := New(mgi)

	err := c.CommentOn(context.Background(), id, "blurb")
	require.Error(t, err)
	assert.Equal(t, code_review.ErrNotFound, err)
}

// Based on a real-world query for a CL that is open and out for review
// with 4 Patchsets
func getOpenChangeInfo() gerrit.ChangeInfo {
	xps := getOpenPatchsets()
	return gerrit.ChangeInfo{
		Id:              "buildbot~master~I29ebaf19a1003e4d9c6df7e5f6469c1f812e0730",
		Created:         time.Date(2019, time.August, 21, 14, 26, 43, 0, time.UTC),
		CreatedString:   "2019-08-21 14:26:43.000000000",
		Updated:         time.Date(2019, time.August, 21, 16, 44, 26, 0, time.UTC),
		UpdatedString:   "2019-08-21 16:44:26.000000000",
		Submitted:       time.Time{},
		SubmittedString: "",
		Project:         "buildbot",
		ChangeId:        "I29ebaf19a1003e4d9c6df7e5f6469c1f812e0730",
		Subject:         "[gold] Add more tryjob processing tests",
		Branch:          "main",
		Committed:       false,
		Revisions: map[string]*gerrit.Revision{
			"337da6ea3a14fd2899b39d0a60c6828971c0d883": xps[3],
			"4cfd5b1ed4d6938efc61fd127bb4a458198ac620": xps[1],
			"787d20c0117d455ef28cce925e2bb5302c2254ad": xps[2],
			"993b807277763b351e72d01e6d65461c4bf57981": xps[0],
		},
		Patchsets:   xps,
		MoreChanges: false,
		Issue:       235460,
		// Labels omitted because it's complex and not needed
		Owner: &gerrit.Person{
			Email: "test@example.com",
		},
		Status:         "NEW",
		WorkInProgress: false,
	}
}

func getOpenPatchsets() []*gerrit.Revision {
	return []*gerrit.Revision{
		{
			ID:            "993b807277763b351e72d01e6d65461c4bf57981",
			Number:        1,
			CreatedString: "2019-08-21 14:26:43.000000000",
			Created:       time.Date(2019, time.August, 21, 14, 26, 43, 0, time.UTC),
			Kind:          "REWORK",
		},
		{
			ID:            "4cfd5b1ed4d6938efc61fd127bb4a458198ac620",
			Number:        2,
			CreatedString: "2019-08-21 15:28:37.000000000",
			Created:       time.Date(2019, time.August, 21, 15, 28, 37, 0, time.UTC),
			Kind:          "REWORK",
		},
		{
			ID:            "787d20c0117d455ef28cce925e2bb5302c2254ad",
			Number:        3,
			CreatedString: "2019-08-21 16:27:34.000000000",
			Created:       time.Date(2019, time.August, 21, 16, 27, 34, 0, time.UTC),
			Kind:          "REWORK",
		},
		{
			ID:            "337da6ea3a14fd2899b39d0a60c6828971c0d883",
			Number:        4,
			CreatedString: "2019-08-21 16:28:38.000000000",
			Created:       time.Date(2019, time.August, 21, 16, 28, 38, 0, time.UTC),
			Kind:          "NO_CODE_CHANGE",
		},
	}
}
