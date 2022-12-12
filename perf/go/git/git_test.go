// Package git is the minimal interface that Perf need to interact with a Git
// repo.
package git

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.skia.org/infra/go/git/testutils"
	"go.skia.org/infra/perf/go/config"
	"go.skia.org/infra/perf/go/git/gittest"
	"go.skia.org/infra/perf/go/types"
)

func TestCockroachDB(t *testing.T) {

	for name, subTest := range subTests {
		t.Run(name, func(t *testing.T) {
			ctx, db, gb, hashes, instanceConfig, cleanup := gittest.NewForTest(t)
			g, err := New(ctx, true, db, instanceConfig)
			require.NoError(t, err)

			subTest(t, ctx, g, gb, hashes, cleanup)
		})
	}
}

// subTestFunction is a func we will call to test one aspect of *SQLTraceStore.
type subTestFunction func(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc)

// subTests are all the tests we have for *SQLTraceStore.
var subTests = map[string]subTestFunction{
	"testDetails_FailOnBadCommitNumber":                                                  testDetails_FailOnBadCommitNumber,
	"testDetails_Success":                                                                testDetails_Success,
	"testCommitSliceFromCommitNumberSlice_EmptyInputSlice_Success":                       testCommitSliceFromCommitNumberSlice_EmptyInputSlice_Success,
	"testCommitSliceFromCommitNumberSlice_Success":                                       testCommitSliceFromCommitNumberSlice_Success,
	"testUpdate_NewCommitsAreFoundAfterUpdate":                                           testUpdate_NewCommitsAreFoundAfterUpdate,
	"testCommitNumberFromGitHash_Success":                                                testCommitNumberFromGitHash_Success,
	"testCommitNumberFromGitHash_ErrorOnUnknownGitHash":                                  testCommitNumberFromGitHash_ErrorOnUnknownGitHash,
	"testCommitNumberFromTime_Success":                                                   testCommitNumberFromTime_Success,
	"testCommitNumberFromTime_ErrorOnTimeTooOld":                                         testCommitNumberFromTime_ErrorOnTimeTooOld,
	"testCommitNumberFromTime_SuccessOnZeroTime":                                         testCommitNumberFromTime_SuccessOnZeroTime,
	"testCommitSliceFromTimeRange_Success":                                               testCommitSliceFromTimeRange_Success,
	"testCommitSliceFromTimeRange_ZeroWidthRangeReturnsZeroResults":                      testCommitSliceFromTimeRange_ZeroWidthRangeReturnsZeroResults,
	"testCommitSliceFromTimeRange_NegativeWidthRangeReturnsZeroResults":                  testCommitSliceFromTimeRange_NegativeWidthRangeReturnsZeroResults,
	"testCommitSliceFromCommitNumberRange_Success":                                       testCommitSliceFromCommitNumberRange_Success,
	"testCommitSliceFromCommitNumberRange_ZeroWidthReturnsOneResult":                     testCommitSliceFromCommitNumberRange_ZeroWidthReturnsOneResult,
	"testCommitSliceFromCommitNumberRange_NegativeWidthReturnsZeroResults":               testCommitSliceFromCommitNumberRange_NegativeWidthReturnsZeroResults,
	"testGitHashFromCommitNumber_Success":                                                testGitHashFromCommitNumber_Success,
	"testGitHashFromCommitNumber_ErrWhenCommitDoesntExist":                               testGitHashFromCommitNumber_ErrWhenCommitDoesntExist,
	"testCommitNumbersWhenFileChangesInCommitNumberRange_Success":                        testCommitNumbersWhenFileChangesInCommitNumberRange_Success,
	"testCommitNumbersWhenFileChangesInCommitNumberRange_EmptySliceIfFileDoesntExist":    testCommitNumbersWhenFileChangesInCommitNumberRange_EmptySliceIfFileDoesntExist,
	"testCommitNumbersWhenFileChangesInCommitNumberRange_RangeIsInclusiveOfBegin":        testCommitNumbersWhenFileChangesInCommitNumberRange_RangeIsInclusiveOfBegin,
	"testCommitNumbersWhenFileChangesInCommitNumberRange_RangeIsInclusiveOfEnd":          testCommitNumbersWhenFileChangesInCommitNumberRange_RangeIsInclusiveOfEnd,
	"testCommitNumbersWhenFileChangesInCommitNumberRange_ResultsWhenBeginEqualsEnd":      testCommitNumbersWhenFileChangesInCommitNumberRange_ResultsWhenBeginEqualsEnd,
	"testCommitNumbersWhenFileChangesInCommitNumberRange_HandlesZeroAsBeginCommitNumber": testCommitNumbersWhenFileChangesInCommitNumberRange_HandlesZeroAsBeginCommitNumber,
	"testLogEntry_Success":                                                               testLogEntry_Success,
	"testLogEntry_BadCommitId_ReturnsError":                                              testLogEntry_BadCommitId_ReturnsError,
}

func testUpdate_NewCommitsAreFoundAfterUpdate(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	// Add new commit to repo, it shouldn't appear in the database.
	newHash := gb.CommitGenAt(ctx, "foo.txt", gittest.StartTime.Add(4*time.Minute))
	_, err := g.CommitNumberFromGitHash(ctx, newHash)
	require.Error(t, err)

	// After Update step we should find it.
	err = g.Update(ctx)
	require.NoError(t, err)
	commitNumber, err := g.CommitNumberFromGitHash(ctx, newHash)
	assert.Equal(t, types.CommitNumber(len(hashes)), commitNumber)
}

func testCommitNumberFromGitHash_Success(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	commitNumber, err := g.CommitNumberFromGitHash(ctx, hashes[0])
	assert.NoError(t, err)
	assert.Equal(t, types.CommitNumber(0), commitNumber)
	commitNumber, err = g.CommitNumberFromGitHash(ctx, hashes[2])
	assert.NoError(t, err)
	assert.Equal(t, types.CommitNumber(2), commitNumber)
}

func testDetails_FailOnBadCommitNumber(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	_, err := g.CommitFromCommitNumber(ctx, types.BadCommitNumber)
	require.Error(t, err)
}

func testDetails_Success(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	commitNumber := types.CommitNumber(1)
	assert.False(t, g.cache.Contains(commitNumber))
	commit, err := g.CommitFromCommitNumber(ctx, commitNumber)
	require.NoError(t, err)

	// The prefix of the URL will change, so just confirm it has the right suffix.
	require.True(t, strings.HasSuffix(commit.URL, commit.GitHash))

	assert.Equal(t, Commit{
		Timestamp:    gittest.StartTime.Add(time.Minute).Unix(),
		GitHash:      hashes[1],
		Author:       "test <test@google.com>",
		Subject:      "501233450539197794",
		URL:          commit.URL,
		CommitNumber: commitNumber,
	}, commit)
	assert.True(t, g.cache.Contains(commitNumber))
}

func testCommitSliceFromCommitNumberSlice_EmptyInputSlice_Success(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	resp, err := g.CommitSliceFromCommitNumberSlice(ctx, []types.CommitNumber{})
	require.NoError(t, err)
	assert.Empty(t, resp)
}

func testCommitSliceFromCommitNumberSlice_Success(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	commitNumbers := []types.CommitNumber{1, 3}
	assert.False(t, g.cache.Contains(commitNumbers[0]))
	assert.False(t, g.cache.Contains(commitNumbers[1]))
	commits, err := g.CommitSliceFromCommitNumberSlice(ctx, commitNumbers)
	require.NoError(t, err)

	// The prefix of the URLs will change, so just confirm it has the right suffix.
	require.True(t, strings.HasSuffix(commits[0].URL, commits[0].GitHash))
	require.True(t, strings.HasSuffix(commits[1].URL, commits[1].GitHash))

	assert.Equal(t, Commit{
		Timestamp:    gittest.StartTime.Add(time.Minute).Unix(),
		GitHash:      hashes[1],
		Author:       "test <test@google.com>",
		Subject:      "501233450539197794",
		URL:          commits[0].URL,
		CommitNumber: commitNumbers[0],
	}, commits[0])
	assert.Equal(t, Commit{
		Timestamp:    gittest.StartTime.Add(3 * time.Minute).Unix(),
		GitHash:      hashes[3],
		Author:       "test <test@google.com>",
		Subject:      "6044372234677422456",
		URL:          commits[1].URL,
		CommitNumber: commitNumbers[1],
	}, commits[1])

	assert.True(t, g.cache.Contains(commitNumbers[0]))
	assert.True(t, g.cache.Contains(commitNumbers[1]))
}

func testCommitNumberFromGitHash_ErrorOnUnknownGitHash(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	_, err := g.CommitNumberFromGitHash(ctx, hashes[0]+"obviously_not_a_valid_hash")
	assert.Error(t, err)
}

func testCommitNumberFromTime_Success(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	// Exact time of commit number 1. See gittest.NewForTest().
	commitNumber, err := g.CommitNumberFromTime(ctx, gittest.StartTime.Add(1*time.Minute))
	assert.NoError(t, err)
	assert.Equal(t, types.CommitNumber(1), commitNumber)

	// A second beyond commit number 1.
	commitNumber, err = g.CommitNumberFromTime(ctx, gittest.StartTime.Add(1*time.Minute+time.Second))
	assert.NoError(t, err)
	assert.Equal(t, types.CommitNumber(1), commitNumber)

	// A second before commit number 1.
	commitNumber, err = g.CommitNumberFromTime(ctx, gittest.StartTime.Add(1*time.Minute-time.Second))
	assert.NoError(t, err)
	assert.Equal(t, types.CommitNumber(0), commitNumber)
}

func testCommitNumberFromTime_SuccessOnZeroTime(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	commitNumber, err := g.CommitNumberFromTime(ctx, time.Time{})
	assert.NoError(t, err)
	assert.Equal(t, types.CommitNumber(len(hashes)-1), commitNumber)
}

func testCommitNumberFromTime_ErrorOnTimeTooOld(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	_, err := g.CommitNumberFromTime(ctx, gittest.StartTime.Add(-1*time.Minute))
	require.Error(t, err)
}

func testCommitSliceFromTimeRange_Success(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	commits, err := g.CommitSliceFromTimeRange(ctx, gittest.StartTime.Add(1*time.Minute), gittest.StartTime.Add(3*time.Minute))
	require.NoError(t, err)
	assert.Len(t, commits, 2)
	assert.Equal(t, int64(1680000060), commits[0].Timestamp)
	assert.Equal(t, types.CommitNumber(1), commits[0].CommitNumber)
	assert.Equal(t, int64(1680000120), commits[1].Timestamp)
	assert.Equal(t, types.CommitNumber(2), commits[1].CommitNumber)
}

func testCommitSliceFromTimeRange_ZeroWidthRangeReturnsZeroResults(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	commits, err := g.CommitSliceFromTimeRange(ctx, gittest.StartTime.Add(1*time.Minute), gittest.StartTime.Add(1*time.Minute))
	require.NoError(t, err)
	assert.Empty(t, commits)
}

func testCommitSliceFromTimeRange_NegativeWidthRangeReturnsZeroResults(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	commits, err := g.CommitSliceFromTimeRange(ctx, gittest.StartTime.Add(2*time.Minute), gittest.StartTime.Add(1*time.Minute))
	require.NoError(t, err)
	assert.Empty(t, commits)
}

func testCommitSliceFromCommitNumberRange_Success(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	commits, err := g.CommitSliceFromCommitNumberRange(ctx, 1, 2)
	require.NoError(t, err)
	require.Len(t, commits, 2)
	assert.Equal(t, int64(1680000060), commits[0].Timestamp)
	assert.Equal(t, types.CommitNumber(1), commits[0].CommitNumber)
	assert.Equal(t, int64(1680000120), commits[1].Timestamp)
	assert.Equal(t, types.CommitNumber(2), commits[1].CommitNumber)
}

func testCommitSliceFromCommitNumberRange_ZeroWidthReturnsOneResult(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	commits, err := g.CommitSliceFromCommitNumberRange(ctx, 2, 2)
	require.NoError(t, err)
	require.Len(t, commits, 1)
	assert.Equal(t, int64(1680000120), commits[0].Timestamp)
	assert.Equal(t, types.CommitNumber(2), commits[0].CommitNumber)
}

func testCommitSliceFromCommitNumberRange_NegativeWidthReturnsZeroResults(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	commits, err := g.CommitSliceFromCommitNumberRange(ctx, 3, 2)
	require.NoError(t, err)
	require.Empty(t, commits)
}

func testGitHashFromCommitNumber_Success(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	gitHash, err := g.GitHashFromCommitNumber(ctx, types.CommitNumber(2))
	require.NoError(t, err)
	assert.Equal(t, hashes[2], gitHash)
}

func testGitHashFromCommitNumber_ErrWhenCommitDoesntExist(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	_, err := g.GitHashFromCommitNumber(ctx, types.BadCommitNumber)
	require.Error(t, err)
}

func testCommitNumbersWhenFileChangesInCommitNumberRange_Success(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	commits, err := g.CommitNumbersWhenFileChangesInCommitNumberRange(ctx, types.CommitNumber(1), types.CommitNumber(7), "bar.txt")
	require.NoError(t, err)
	assert.Equal(t, []types.CommitNumber{3, 6}, commits)
}

func testCommitNumbersWhenFileChangesInCommitNumberRange_EmptySliceIfFileDoesntExist(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	commits, err := g.CommitNumbersWhenFileChangesInCommitNumberRange(ctx, types.CommitNumber(1), types.CommitNumber(7), "this-file-doesnt-exist.txt")
	require.NoError(t, err)
	assert.Empty(t, commits)
}

func testCommitNumbersWhenFileChangesInCommitNumberRange_RangeIsInclusiveOfBegin(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	commits, err := g.CommitNumbersWhenFileChangesInCommitNumberRange(ctx, types.CommitNumber(3), types.CommitNumber(7), "bar.txt")
	require.NoError(t, err)
	assert.Equal(t, []types.CommitNumber{3, 6}, commits)
}

func testCommitNumbersWhenFileChangesInCommitNumberRange_RangeIsInclusiveOfEnd(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	commits, err := g.CommitNumbersWhenFileChangesInCommitNumberRange(ctx, types.CommitNumber(5), types.CommitNumber(6), "bar.txt")
	require.NoError(t, err)
	assert.Equal(t, []types.CommitNumber{6}, commits)
}

func testCommitNumbersWhenFileChangesInCommitNumberRange_ResultsWhenBeginEqualsEnd(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	commits, err := g.CommitNumbersWhenFileChangesInCommitNumberRange(ctx, types.CommitNumber(6), types.CommitNumber(6), "bar.txt")
	require.NoError(t, err)
	assert.Equal(t, []types.CommitNumber{6}, commits)
}

func testCommitNumbersWhenFileChangesInCommitNumberRange_HandlesZeroAsBeginCommitNumber(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	commits, err := g.CommitNumbersWhenFileChangesInCommitNumberRange(ctx, types.CommitNumber(0), types.CommitNumber(4), "bar.txt")
	require.NoError(t, err)
	assert.Equal(t, []types.CommitNumber{3}, commits)
}

func testLogEntry_Success(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	got, err := g.LogEntry(ctx, types.CommitNumber(1))
	require.NoError(t, err)
	expected := `commit 881dfc43620250859549bb7e0301b6910d9b8e70
Author: test <test@google.com>
Date:   Tue Mar 28 10:41:00 2023 +0000

    501233450539197794
`
	require.Equal(t, expected, got)
}

func testLogEntry_BadCommitId_ReturnsError(t *testing.T, ctx context.Context, g *Git, gb *testutils.GitBuilder, hashes []string, cleanup gittest.CleanupFunc) {
	defer cleanup()

	_, err := g.LogEntry(ctx, types.BadCommitNumber)
	require.Error(t, err)
}

func TestParseGitRevLogStream_Success(t *testing.T) {
	r := strings.NewReader(
		`commit 6079a7810530025d9877916895dd14eb8bb454c0
Joe Gregorio <joe@bitworking.org>
Change #9
1584837783`)

	err := parseGitRevLogStream(ioutil.NopCloser(r), func(p Commit) error {
		assert.Equal(t, Commit{
			CommitNumber: types.BadCommitNumber,
			GitHash:      "6079a7810530025d9877916895dd14eb8bb454c0",
			Timestamp:    1584837783,
			Author:       "Joe Gregorio <joe@bitworking.org>",
			Subject:      "Change #9"}, p)
		return nil
	})
	assert.NoError(t, err)
}

func TestParseGitRevLogStream_ErrPropagatesWhenCallbackReturnsError(t *testing.T) {
	r := strings.NewReader(
		`commit 6079a7810530025d9877916895dd14eb8bb454c0
Joe Gregorio <joe@bitworking.org>
Change #9
1584837783`)

	err := parseGitRevLogStream(ioutil.NopCloser(r), func(p Commit) error {
		return fmt.Errorf("This is an error.")
	})
	assert.Contains(t, err.Error(), "This is an error.")
}

func TestParseGitRevLogStream_SuccessForTwoCommits(t *testing.T) {
	r := strings.NewReader(
		`commit 6079a7810530025d9877916895dd14eb8bb454c0
Joe Gregorio <joe@bitworking.org>
Change #9
1584837783
commit 977e0ef44bec17659faf8c5d4025c5a068354817
Joe Gregorio <joe@bitworking.org>
Change #8
1584837780`)
	count := 0
	hashes := []string{"6079a7810530025d9877916895dd14eb8bb454c0", "977e0ef44bec17659faf8c5d4025c5a068354817"}
	err := parseGitRevLogStream(ioutil.NopCloser(r), func(p Commit) error {
		assert.Equal(t, "Joe Gregorio <joe@bitworking.org>", p.Author)
		assert.Equal(t, hashes[count], p.GitHash)
		count++
		return nil
	})
	assert.Equal(t, 2, count)
	assert.NoError(t, err)
}

func TestParseGitRevLogStream_EmptyFile_Success(t *testing.T) {
	r := strings.NewReader("")
	err := parseGitRevLogStream(ioutil.NopCloser(r), func(p Commit) error {
		assert.Fail(t, "Should never get here.")
		return nil
	})
	assert.NoError(t, err)
}

func TestParseGitRevLogStream_ErrMissingTimestamp(t *testing.T) {
	r := strings.NewReader(
		`commit 6079a7810530025d9877916895dd14eb8bb454c0
Joe Gregorio <joe@bitworking.org>
Change #9`)
	err := parseGitRevLogStream(ioutil.NopCloser(r), func(p Commit) error {
		assert.Fail(t, "Should never get here.")
		return nil
	})
	assert.Contains(t, err.Error(), "expecting a timestamp")
}

func TestParseGitRevLogStream_ErrFailedToParseTimestamp(t *testing.T) {
	r := strings.NewReader(
		`commit 6079a7810530025d9877916895dd14eb8bb454c0
Joe Gregorio <joe@bitworking.org>
Change #9
ooops 1584837780`)
	err := parseGitRevLogStream(ioutil.NopCloser(r), func(p Commit) error {
		assert.Fail(t, "Should never get here.")
		return nil
	})
	assert.Contains(t, err.Error(), "Failed to parse timestamp")
}

func TestParseGitRevLogStream_ErrMissingSubject(t *testing.T) {
	r := strings.NewReader(
		`commit 6079a7810530025d9877916895dd14eb8bb454c0
Joe Gregorio <joe@bitworking.org>`)
	err := parseGitRevLogStream(ioutil.NopCloser(r), func(p Commit) error {
		assert.Fail(t, "Should never get here.")
		return nil
	})
	assert.Contains(t, err.Error(), "expecting a subject")
}

func TestParseGitRevLogStream_ErrMissingAuthor(t *testing.T) {
	r := strings.NewReader(
		`commit 6079a7810530025d9877916895dd14eb8bb454c0`)
	err := parseGitRevLogStream(ioutil.NopCloser(r), func(p Commit) error {
		assert.Fail(t, "Should never get here.")
		return nil
	})
	assert.Contains(t, err.Error(), "expecting an author")
}

func TestParseGitRevLogStream_ErrMalformedCommitLine(t *testing.T) {
	r := strings.NewReader(
		`something_not_commit 6079a7810530025d9877916895dd14eb8bb454c0`)
	err := parseGitRevLogStream(ioutil.NopCloser(r), func(p Commit) error {
		assert.Fail(t, "Should never get here.")
		return nil
	})
	assert.Contains(t, err.Error(), "expected commit at")
}

func TestURLFromParts_DebounceCommitURL_Success(t *testing.T) {

	const debounceURL = "https://some.other.url.example.org"
	instanceConfig := &config.InstanceConfig{
		GitRepoConfig: config.GitRepoConfig{
			URL:              "https://skia.googlesource.com/skia",
			DebouceCommitURL: true,
		},
	}
	commit := Commit{
		GitHash: "6079a7810530025d9877916895dd14eb8bb454c0",
		Subject: debounceURL,
	}
	assert.Equal(t, debounceURL, urlFromParts(instanceConfig, commit))
}

func TestURLFromParts_CommitURLSupplied_Success(t *testing.T) {

	instanceConfig := &config.InstanceConfig{
		GitRepoConfig: config.GitRepoConfig{
			URL:       "https://github.com/google/skia",
			CommitURL: "%s/commit/%s",
		},
	}
	commit := Commit{
		GitHash: "6079a7810530025d9877916895dd14eb8bb454c0",
	}
	assert.Equal(t, "https://github.com/google/skia/commit/6079a7810530025d9877916895dd14eb8bb454c0", urlFromParts(instanceConfig, commit))
}

func TestURLFromParts_DefaultCommitURL_Success(t *testing.T) {

	instanceConfig := &config.InstanceConfig{
		GitRepoConfig: config.GitRepoConfig{
			URL: "https://skia.googlesource.com/skia",
		},
	}
	commit := Commit{
		GitHash: "6079a7810530025d9877916895dd14eb8bb454c0",
	}
	assert.Equal(t, "https://skia.googlesource.com/skia/+show/6079a7810530025d9877916895dd14eb8bb454c0", urlFromParts(instanceConfig, commit))
}

func TestCommit_Display(t *testing.T) {

	c := Commit{
		CommitNumber: 10223,
		GitHash:      "d261e1075a93677442fdf7fe72aba7e583863664",
		Timestamp:    1498176000,
		Author:       "Robert Phillips <robertphillips@google.com>",
		Subject:      "Re-enable opList dependency tracking",
		URL:          "https://skia.googlesource.com/skia/+show/d261e1075a93677442fdf7fe72aba7e583863664",
	}
	assert.Equal(t, "d261e10 -  2y 40w - Re-enable opList dependency tracking", c.Display(time.Date(2020, 04, 01, 0, 0, 0, 0, time.UTC)))
}
