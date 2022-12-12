package dfbuilder

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opencensus.io/trace"
	"go.skia.org/infra/go/metrics2"
	"go.skia.org/infra/go/paramtools"
	"go.skia.org/infra/go/query"
	"go.skia.org/infra/go/skerr"
	"go.skia.org/infra/go/sklog"
	"go.skia.org/infra/go/timer"
	"go.skia.org/infra/go/vec32"
	"go.skia.org/infra/perf/go/dataframe"
	perfgit "go.skia.org/infra/perf/go/git"
	"go.skia.org/infra/perf/go/progress"
	"go.skia.org/infra/perf/go/tracesetbuilder"
	"go.skia.org/infra/perf/go/tracestore"
	"go.skia.org/infra/perf/go/types"
	"golang.org/x/sync/errgroup"
)

const (
	// newNMaxSearch is the minimum number of queries to perform that returned
	// no data before giving up.
	// TODO(jcgregorio) Make this either a flag or config value.
	newNMaxSearch = 40

	// It is possible for some ParamSet tiles to have "bad" data, for example, a
	// tremendous amount of garbage data from a bad ingestion process. This
	// timeout guards against that eventuality by limiting how long Queries get
	// to run on each tile.
	//
	// Note that this value is much shorter than the default timeout for all SQL
	// requests because during regression detection we can issue many Queries and
	// as each one hits that bad ParamSet tile we can clog the backend with these
	// long running queries.
	singleTileQueryTimeout = time.Minute
)

// builder implements DataFrameBuilder using TraceStore.
type builder struct {
	git               *perfgit.Git
	store             tracestore.TraceStore
	tileSize          int32
	numPreflightTiles int

	newTimer                      metrics2.Float64SummaryMetric
	newByTileTimer                metrics2.Float64SummaryMetric
	newFromQueryAndRangeTimer     metrics2.Float64SummaryMetric
	newFromKeysAndRangeTimer      metrics2.Float64SummaryMetric
	newFromCommitIDsAndQueryTimer metrics2.Float64SummaryMetric
	newNFromQueryTimer            metrics2.Float64SummaryMetric
	newNFromKeysTimer             metrics2.Float64SummaryMetric
	preflightQueryTimer           metrics2.Float64SummaryMetric
}

// NewDataFrameBuilderFromTraceStore builds a DataFrameBuilder.
func NewDataFrameBuilderFromTraceStore(git *perfgit.Git, store tracestore.TraceStore, numPreflightTiles int) dataframe.DataFrameBuilder {
	return &builder{
		git:                           git,
		store:                         store,
		numPreflightTiles:             numPreflightTiles,
		tileSize:                      store.TileSize(),
		newTimer:                      metrics2.GetFloat64SummaryMetric("perfserver_dfbuilder_new"),
		newByTileTimer:                metrics2.GetFloat64SummaryMetric("perfserver_dfbuilder_newByTile"),
		newFromQueryAndRangeTimer:     metrics2.GetFloat64SummaryMetric("perfserver_dfbuilder_newFromQueryAndRange"),
		newFromKeysAndRangeTimer:      metrics2.GetFloat64SummaryMetric("perfserver_dfbuilder_newFromKeysAndRange"),
		newFromCommitIDsAndQueryTimer: metrics2.GetFloat64SummaryMetric("perfserver_dfbuilder_newFromCommitIDsAndQuery"),
		newNFromQueryTimer:            metrics2.GetFloat64SummaryMetric("perfserver_dfbuilder_newNFromQuery"),
		newNFromKeysTimer:             metrics2.GetFloat64SummaryMetric("perfserver_dfbuilder_newNFromKeys"),
		preflightQueryTimer:           metrics2.GetFloat64SummaryMetric("perfserver_dfbuilder_preflightQuery"),
	}
}

// fromIndexRange returns the headers and indices for all the commits
// between beginIndex and endIndex inclusive.
func fromIndexRange(ctx context.Context, git *perfgit.Git, beginIndex, endIndex types.CommitNumber) ([]*dataframe.ColumnHeader, []types.CommitNumber, int, error) {
	ctx, span := trace.StartSpan(ctx, "dfbuilder.fromIndexRange")
	defer span.End()

	commits, err := git.CommitSliceFromCommitNumberRange(ctx, beginIndex, endIndex)
	if err != nil {
		return nil, nil, 0, skerr.Wrapf(err, "Failed to get headers and commit numbers from time range.")
	}
	colHeader := make([]*dataframe.ColumnHeader, len(commits), len(commits))
	commitNumbers := make([]types.CommitNumber, len(commits), len(commits))
	for i, commit := range commits {
		colHeader[i] = &dataframe.ColumnHeader{
			Offset:    commit.CommitNumber,
			Timestamp: commit.Timestamp,
		}
		commitNumbers[i] = commit.CommitNumber
	}
	return colHeader, commitNumbers, 0, nil
}

// tileMapOffsetToIndex maps the offset of each point in a tile to the index it
// should appear in the resulting Trace.
type tileMapOffsetToIndex map[types.TileNumber]map[int32]int32

// buildTileMapOffsetToIndex returns a tileMapOffsetToIndex for the given indices and the given TraceStore.
//
// The returned map is used when loading traces out of tiles.
func buildTileMapOffsetToIndex(indices []types.CommitNumber, store tracestore.TraceStore) tileMapOffsetToIndex {
	ret := tileMapOffsetToIndex{}
	for targetIndex, commitNumber := range indices {
		tileNumber := store.TileNumber(commitNumber)
		if traceMap, ok := ret[tileNumber]; !ok {
			ret[tileNumber] = map[int32]int32{
				store.OffsetFromCommitNumber(commitNumber): int32(targetIndex),
			}
		} else {
			traceMap[store.OffsetFromCommitNumber(commitNumber)] = int32(targetIndex)
		}
	}
	return ret
}

// new builds a DataFrame for the given columns and populates it with traces that match the given query.
//
// The progress callback is triggered once for every tile.
func (b *builder) new(ctx context.Context, colHeaders []*dataframe.ColumnHeader, indices []types.CommitNumber, q *query.Query, progress progress.Progress, skip int) (*dataframe.DataFrame, error) {
	ctx, span := trace.StartSpan(ctx, "dfbuilder.new")
	defer span.End()

	// TODO tickle progress as each Go routine completes.
	defer timer.NewWithSummary("perfserver_dfbuilder_new", b.newTimer).Stop()
	// Determine which tiles we are querying over, and how each tile maps into our results.
	mapper := buildTileMapOffsetToIndex(indices, b.store)

	traceSetBuilder := tracesetbuilder.New(len(indices))
	defer traceSetBuilder.Close()

	var mutex sync.Mutex // mutex protects stepsCompleted.
	stepsCompleted := 0
	triggerProgress := func() {
		mutex.Lock()
		defer mutex.Unlock()
		stepsCompleted += 1
		progress.Message("Step", fmt.Sprintf("%d/%d", stepsCompleted, len(mapper)))
	}

	var g errgroup.Group
	// For each tile.
	for tileNumber, traceMap := range mapper {
		traceMap := traceMap
		tileNumber := tileNumber
		// TODO(jcgregorio) If we query across a large number of tiles N then this will spawn N*8 Go routines
		// all hitting the backend at the same time. Maybe we need a worker pool if this becomes a problem.
		g.Go(func() error {
			defer timer.NewWithSummary("perfserver_dfbuilder_new_by_tile", b.newByTileTimer).Stop()

			// Query for matching traces in the given tile.
			queryContext, cancel := context.WithTimeout(ctx, singleTileQueryTimeout)
			defer cancel()
			traces, err := b.store.QueryTraces(queryContext, tileNumber, q)
			if err != nil {
				return err
			}

			traceSetBuilder.Add(traceMap, traces)
			triggerProgress()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		span.SetStatus(trace.Status{
			Code:    trace.StatusCodeInternal,
			Message: err.Error(),
		})

		return nil, fmt.Errorf("Failed while querying: %s", err)
	}
	traceSet, paramSet := traceSetBuilder.Build(ctx)
	d := &dataframe.DataFrame{
		TraceSet: traceSet,
		Header:   colHeaders,
		ParamSet: paramSet,
		Skip:     skip,
	}
	return d, nil
}

// See DataFrameBuilder.
func (b *builder) NewFromQueryAndRange(ctx context.Context, begin, end time.Time, q *query.Query, downsample bool, progress progress.Progress) (*dataframe.DataFrame, error) {
	ctx, span := trace.StartSpan(ctx, "dfbuilder.NewFromQueryAndRange")
	defer span.End()

	defer timer.NewWithSummary("perfserver_dfbuilder_NewFromQueryAndRange", b.newFromQueryAndRangeTimer).Stop()

	colHeaders, indices, skip, err := dataframe.FromTimeRange(ctx, b.git, begin, end, downsample)
	if err != nil {
		return nil, skerr.Wrap(err)
	}
	return b.new(ctx, colHeaders, indices, q, progress, skip)
}

// See DataFrameBuilder.
func (b *builder) NewFromKeysAndRange(ctx context.Context, keys []string, begin, end time.Time, downsample bool, progress progress.Progress) (*dataframe.DataFrame, error) {
	ctx, span := trace.StartSpan(ctx, "dfbuilder.NewFromKeysAndRange")
	defer span.End()

	// TODO tickle progress as each Go routine completes.
	defer timer.NewWithSummary("perfserver_dfbuilder_NewFromKeysAndRange", b.newFromKeysAndRangeTimer).Stop()
	colHeaders, indices, skip, err := dataframe.FromTimeRange(ctx, b.git, begin, end, downsample)
	if err != nil {
		return nil, skerr.Wrap(err)
	}

	// Determine which tiles we are querying over, and how each tile maps into our results.
	mapper := buildTileMapOffsetToIndex(indices, b.store)

	var mutex sync.Mutex // mutex protects traceSet and paramSet.
	traceSet := types.TraceSet{}
	paramSet := paramtools.ParamSet{}
	stepsCompleted := 0
	// triggerProgress must only be called when the caller has mutex locked.
	triggerProgress := func() {
		stepsCompleted += 1
		progress.Message("Step", fmt.Sprintf("%d/%d", stepsCompleted, len(mapper)))
	}

	var g errgroup.Group
	// For each tile.
	for tileKey, traceMap := range mapper {
		tileKey := tileKey
		traceMap := traceMap
		g.Go(func() error {
			// Read the traces for the given keys.
			traces, err := b.store.ReadTraces(ctx, tileKey, keys)
			if err != nil {
				return err
			}
			mutex.Lock()
			defer mutex.Unlock()
			// For each trace, convert the encodedKey to a structured key
			// and copy the trace values into their final destination.
			for key, tileTrace := range traces {
				trace, ok := traceSet[key]
				if !ok {
					trace = types.NewTrace(len(indices))
				}
				for srcIndex, dstIndex := range traceMap {
					trace[dstIndex] = tileTrace[srcIndex]
				}
				traceSet[key] = trace
				p, err := query.ParseKey(key)
				if err != nil {
					continue
				}
				paramSet.AddParams(p)
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("Failed while querying: %s", err)
	}
	d := &dataframe.DataFrame{
		TraceSet: traceSet,
		Header:   colHeaders,
		ParamSet: paramSet.Freeze(),
		Skip:     skip,
	}
	triggerProgress()
	return d, nil
}

// findIndexForTime finds the index of the closest commit <= 'end'.
//
// Pass in zero time, i.e. time.Time{} to indicate to just get the most recent commit.
func (b *builder) findIndexForTime(ctx context.Context, end time.Time) (types.CommitNumber, error) {
	ctx, span := trace.StartSpan(ctx, "dfbuilder.findIndexForTime")
	defer span.End()

	return b.git.CommitNumberFromTime(ctx, end)
}

// See DataFrameBuilder.
func (b *builder) NewNFromQuery(ctx context.Context, end time.Time, q *query.Query, n int32, progress progress.Progress) (*dataframe.DataFrame, error) {
	ctx, span := trace.StartSpan(ctx, "dfbuilder.NewNFromQuery")
	defer span.End()

	defer timer.NewWithSummary("perfserver_dfbuilder_NewNFromQuery", b.newNFromQueryTimer).Stop()

	sklog.Infof("Querying to: %v", end)

	ret := dataframe.NewEmpty()
	ps := paramtools.NewParamSet()
	var total int32 // total number of commits we've added to ret so far.
	steps := 1      // Total number of times we've gone through the loop below, used in the progress() callback.
	numStepsNoData := 0

	endIndex, err := b.findIndexForTime(ctx, end)
	if err != nil {
		return nil, fmt.Errorf("Failed to find end index: %s", err)
	}
	// beginIndex is the index of the first commit in the tile that endIndex is
	// in. We are OK if beginIndex == endIndex because fromIndexRange returns
	// headers from begin to end *inclusive*.
	beginIndex := b.store.CommitNumberOfTileStart(endIndex)

	sklog.Infof("BeginIndex: %d  EndIndex: %d", beginIndex, endIndex)
	for total < n {

		// Query for traces.
		headers, indices, skip, err := fromIndexRange(ctx, b.git, beginIndex, endIndex)
		if err != nil {
			return nil, fmt.Errorf("Failed building index range: %s", err)
		}

		df, err := b.new(ctx, headers, indices, q, progress, skip)
		if err != nil {
			return nil, fmt.Errorf("Failed while querying: %s", err)
		}

		nonMissing := 0
		// Total up the number of data points we have for each commit.
		counts := make([]int, len(df.Header))
		for _, tr := range df.TraceSet {
			for i, x := range tr {
				if x != vec32.MissingDataSentinel {
					counts[i] += 1
					nonMissing += 1
				}
			}
		}
		// If there are no matches then we might be done.
		if nonMissing == 0 {
			numStepsNoData += 1
		}
		if numStepsNoData > newNMaxSearch {
			sklog.Infof("Failed querying: %s", q)
			break
		}

		ps.AddParamSet(df.ParamSet)

		// For each commit that has data, copy the data from df into ret.
		// Move backwards down the trace since we are building the result from 'end' backwards.
		for i := len(counts) - 1; i >= 0; i-- {
			if counts[i] > 0 {
				ret.Header = append([]*dataframe.ColumnHeader{df.Header[i]}, ret.Header...)
				for key, sourceTrace := range df.TraceSet {
					if _, ok := ret.TraceSet[key]; !ok {
						ret.TraceSet[key] = vec32.New(int(n))
					}
					ret.TraceSet[key][n-1-total] = sourceTrace[i]
				}
				total += 1

				// If we've added enough commits to ret then we are done.
				if total == n {
					break
				}
			}
		}
		sklog.Infof("Total: %d Steps: %d NumStepsNoData: %d Query: %v", total, steps, numStepsNoData, q.String())

		if total == n {
			break
		}

		progress.Message("Tiles", fmt.Sprintf("Tiles searched: %d. Found %d/%d points.", steps, total, n))
		steps += 1

		// Now step back a full tile.

		// At this point we know beginIndex points to the 0th column in a tile,
		// so endIndex is easy to calculate.
		endIndex = beginIndex - 1
		if endIndex < 0 {
			break
		}
		beginIndex = b.store.CommitNumberOfTileStart(endIndex)
		if beginIndex < 0 {
			beginIndex = 0
		}
	}
	ps.Normalize()
	ret.ParamSet = ps.Freeze()

	if total < n {
		// Trim down the traces so they are the same length as ret.Header.
		for key, tr := range ret.TraceSet {
			ret.TraceSet[key] = tr[n-total:]
		}
	}

	return ret, nil
}

// See DataFrameBuilder.
func (b *builder) NewNFromKeys(ctx context.Context, end time.Time, keys []string, n int32, progress progress.Progress) (*dataframe.DataFrame, error) {
	ctx, span := trace.StartSpan(ctx, "dfbuilder.NewNFromKeys")
	defer span.End()

	defer timer.NewWithSummary("perfserver_dfbuilder_NewNFromKeys", b.newNFromKeysTimer).Stop()

	endIndex, err := b.findIndexForTime(ctx, end)
	if err != nil {
		return nil, fmt.Errorf("Failed to find end index: %s", err)
	}
	beginIndex := endIndex.Add(-(b.tileSize - 1))
	if beginIndex < 0 {
		beginIndex = 0
	}

	ret := dataframe.NewEmpty()
	ps := paramtools.NewParamSet()
	var total int32 // total number of commits we've added to ret so far.
	steps := 1      // Total number of times we've gone through the loop below, used in the progress() callback.
	numStepsNoData := 0

	for total < n {
		headers, indices, skip, err := fromIndexRange(ctx, b.git, beginIndex, endIndex)
		if err != nil {
			return nil, fmt.Errorf("Failed building index range: %s", err)
		}

		// Determine which tiles we are querying over, and how each tile maps into our results.
		mapper := buildTileMapOffsetToIndex(indices, b.store)

		traceSet := types.TraceSet{}
		for tileKey, traceMap := range mapper {
			// Read the traces for the given keys.
			traces, err := b.store.ReadTraces(ctx, tileKey, keys)
			if err != nil {
				return nil, err
			}
			// For each trace, convert the encodedKey to a structured key
			// and copy the trace values into their final destination.
			for key, tileTrace := range traces {
				trace, ok := traceSet[key]
				if !ok {
					trace = types.NewTrace(len(indices))
				}
				for srcIndex, dstIndex := range traceMap {
					trace[dstIndex] = tileTrace[srcIndex]
				}
				traceSet[key] = trace
			}
		}
		df := &dataframe.DataFrame{
			TraceSet: traceSet,
			Header:   headers,
			ParamSet: paramtools.NewReadOnlyParamSet(),
			Skip:     skip,
		}
		df.BuildParamSet()

		nonMissing := 0
		// Total up the number of data points we have for each commit.
		counts := make([]int, len(df.Header))
		for _, tr := range df.TraceSet {
			for i, x := range tr {
				if x != vec32.MissingDataSentinel {
					counts[i] += 1
					nonMissing += 1
				}
			}
		}
		// If there are no matches then we might be done.
		if nonMissing == 0 {
			numStepsNoData += 1
		}
		if numStepsNoData > newNMaxSearch {
			break
		}

		ps.AddParamSet(df.ParamSet)

		// For each commit that has data, copy the data from df into ret.
		// Move backwards down the trace since we are building the result from 'end' backwards.
		for i := len(counts) - 1; i >= 0; i-- {
			if counts[i] > 0 {
				ret.Header = append([]*dataframe.ColumnHeader{df.Header[i]}, ret.Header...)
				for key, sourceTrace := range df.TraceSet {
					if _, ok := ret.TraceSet[key]; !ok {
						ret.TraceSet[key] = vec32.New(int(n))
					}
					ret.TraceSet[key][n-1-total] = sourceTrace[i]
				}
				total += 1

				// If we've added enough commits to ret then we are done.
				if total == n {
					break
				}
			}
		}

		sklog.Infof("Total: %d Steps: %d NumStepsNoData: %d", total, steps, numStepsNoData)

		if total == n {
			break
		}

		progress.Message("Tiles", fmt.Sprintf("Tiles searched: %d. Found %d/%d points.", steps, total, n))
		steps += 1

		endIndex = endIndex.Add(-b.tileSize)
		beginIndex = endIndex.Add(-b.tileSize)
		if endIndex < 0 {
			break
		}
		if beginIndex < 0 {
			beginIndex = 0
		}
	}
	ps.Normalize()
	ret.ParamSet = ps.Freeze()

	if total < n {
		// Trim down the traces so they are the same length as ret.Header.
		for key, tr := range ret.TraceSet {
			ret.TraceSet[key] = tr[n-total:]
		}
	}

	return ret, nil
}

// PreflightQuery implements dataframe.DataFrameBuilder.
func (b *builder) PreflightQuery(ctx context.Context, q *query.Query, referenceParamSet paramtools.ReadOnlyParamSet) (int64, paramtools.ParamSet, error) {
	ctx, span := trace.StartSpan(ctx, "dfbuiler.PreflightQuery")
	defer span.End()

	defer timer.NewWithSummary("perfserver_dfbuilder_PreflightQuery", b.preflightQueryTimer).Stop()

	var count int64

	tileNumber, err := b.store.GetLatestTile(ctx)
	if err != nil {
		return -1, nil, err
	}

	if q.Empty() {
		return -1, nil, skerr.Fmt("Can not pre-flight an empty query")
	}

	// Since the query isn't empty we'll have to run a partial query
	// to build the ParamSet. Do so over the two most recent tiles.
	ps := paramtools.NewParamSet()

	queryContext, cancel := context.WithTimeout(ctx, time.Duration(b.numPreflightTiles)*singleTileQueryTimeout)
	defer cancel()
	for i := 0; i < b.numPreflightTiles; i++ {
		// Count the matches and sum the params in the tile.
		out, err := b.store.QueryTracesIDOnly(queryContext, tileNumber, q)
		if err != nil {
			return -1, nil, fmt.Errorf("failed to query traces: %s", err)
		}
		var tileOneCount int64
		for p := range out {
			tileOneCount++
			ps.AddParams(p)
		}
		if tileOneCount > count {
			count = tileOneCount
		}
		// Now move to the previous tile.
		tileNumber = tileNumber.Prev()
		if tileNumber == types.BadTileNumber {
			break
		}
	}

	// Now we have the ParamSet that corresponds to the query, but for each
	// key in the query we need to go back and put in all the values that
	// appear for that key since the user can make more selections in that
	// key.
	queryPlan, err := q.QueryPlan(ps.Freeze())
	if err != nil {
		return -1, nil, err
	}
	for key := range queryPlan {
		ps[key] = referenceParamSet[key]
	}
	ps.Normalize()

	return count, ps, nil
}

// NumMatches implements dataframe.DataFrameBuilder.
func (b *builder) NumMatches(ctx context.Context, q *query.Query) (int64, error) {
	ctx, span := trace.StartSpan(ctx, "dfbuiler.NumMatches")
	defer span.End()
	defer timer.NewWithSummary("perfserver_dfbuilder_NumMatches", b.preflightQueryTimer).Stop()

	var count int64

	tileNumber, err := b.store.GetLatestTile(ctx)
	if err != nil {
		return -1, skerr.Wrap(err)
	}

	queryContext, cancel := context.WithTimeout(ctx, 2*singleTileQueryTimeout)
	defer cancel()

	// Count the matches in the first tile.
	out, err := b.store.QueryTracesIDOnly(queryContext, tileNumber, q)
	if err != nil {
		return -1, skerr.Wrapf(err, "Failed to query traces.")
	}
	var tileOneCount int64
	for range out {
		tileOneCount++
	}
	count = tileOneCount

	// Now move to the previous tile.
	tileNumber = tileNumber.Prev()
	if tileNumber != types.BadTileNumber {
		// Count the matches in the second tile.
		out, err = b.store.QueryTracesIDOnly(queryContext, tileNumber, q)
		if err != nil {
			return -1, fmt.Errorf("Failed to query traces: %s", err)
		}
		var tileTwoCount int64
		for range out {
			tileTwoCount++
		}
		// Use the larger of the two counts as our result.
		if tileTwoCount > count {
			count = tileTwoCount
		}
	}

	return count, nil
}

// Validate that *builder faithfully implements the DataFrameBuidler interface.
var _ dataframe.DataFrameBuilder = (*builder)(nil)
