package clustering2

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.skia.org/infra/go/now"
	"go.skia.org/infra/go/paramtools"
	"go.skia.org/infra/perf/go/ctrace2"
	"go.skia.org/infra/perf/go/dataframe"
	"go.skia.org/infra/perf/go/kmeans"
	"go.skia.org/infra/perf/go/types"
)

func TestNewClusterSummary_RecordsTheTimeTheClusterSummaryWasCreated_Success(t *testing.T) {

	testTime := time.Date(2020, 05, 01, 12, 00, 00, 00, time.UTC)
	ctx := context.WithValue(context.Background(), now.ContextKey, testTime)
	cs := NewClusterSummary(ctx)
	require.Equal(t, testTime, cs.Timestamp)
}

func TestParamSummaries(t *testing.T) {
	obs := []kmeans.Clusterable{
		ctrace2.NewFullTrace(",arch=x86,config=8888,", []float32{1, 2}, 0.001),
		ctrace2.NewFullTrace(",arch=x86,config=565,", []float32{2, 3}, 0.001),
		ctrace2.NewFullTrace(",arch=x86,config=565,", []float32{3, 2}, 0.001),
	}
	expected := []ValuePercent{
		{"arch=x86", 100},
		{"config=565", 66},
		{"config=8888", 33},
	}
	assert.Equal(t, expected, getParamSummaries(obs))

	obs = []kmeans.Clusterable{}
	expected = []ValuePercent{}
	assert.Equal(t, expected, getParamSummaries(obs))
}

func TestCalcCusterSummaries(t *testing.T) {

	ctx := context.Background()
	rand.Seed(1)
	now := time.Now()
	df := &dataframe.DataFrame{
		TraceSet: types.TraceSet{
			",arch=x86,config=8888,": []float32{0, 0, 1, 1, 1},
			",arch=x86,config=565,":  []float32{0, 0, 1, 1, 1},
			",arch=arm,config=8888,": []float32{1, 1, 1, 1, 1},
			",arch=arm,config=565,":  []float32{1, 1, 1, 1, 1},
		},
		Header: []*dataframe.ColumnHeader{
			{
				Offset:    0,
				Timestamp: now.Unix(),
			},
			{
				Offset:    1,
				Timestamp: now.Add(time.Minute).Unix(),
			},
			{
				Offset:    2,
				Timestamp: now.Add(2 * time.Minute).Unix(),
			},
			{
				Offset:    3,
				Timestamp: now.Add(3 * time.Minute).Unix(),
			},
			{
				Offset:    4,
				Timestamp: now.Add(4 * time.Minute).Unix(),
			},
		},
		ParamSet: paramtools.NewReadOnlyParamSet(),
		Skip:     0,
	}
	ps := paramtools.NewParamSet()
	for key := range df.TraceSet {
		ps.AddParamsFromKey(key)
	}
	df.ParamSet = ps.Freeze()
	sum, err := CalculateClusterSummaries(ctx, df, 4, 0.01, nil, 50, types.OriginalStep)
	assert.NoError(t, err)
	assert.NotNil(t, sum)
	assert.Equal(t, 2, len(sum.Clusters))
	assert.Equal(t, df.Header[2], sum.Clusters[0].StepPoint)
	assert.Equal(t, 2, len(sum.Clusters[0].Keys))
	assert.Equal(t, 2, len(sum.Clusters[1].Keys))
}

func TestCalcCusterSummariesDegenerate(t *testing.T) {

	ctx := context.Background()
	rand.Seed(1)
	df := &dataframe.DataFrame{
		TraceSet: types.TraceSet{},
		Header:   []*dataframe.ColumnHeader{},
		ParamSet: paramtools.NewReadOnlyParamSet(),
		Skip:     0,
	}
	_, err := CalculateClusterSummaries(ctx, df, 4, 0.01, nil, 50, types.OriginalStep)
	assert.Error(t, err)
}
