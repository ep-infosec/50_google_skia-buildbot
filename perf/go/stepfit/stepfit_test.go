package stepfit

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.skia.org/infra/go/vec32"
	"go.skia.org/infra/perf/go/types"
)

const (
	// x values are supplied but ignored in all the tests below.
	x = vec32.MissingDataSentinel

	// minStdDev is the minimum standard deviation used in all tests.
	minStdDev = 0.1
)

func TestStepFit_Empty(t *testing.T) {
	assert.Equal(t,
		&StepFit{TurningPoint: 0, StepSize: 0, Status: UNINTERESTING, Regression: 0},
		GetStepFitAtMid([]float32{}, minStdDev, 20, types.OriginalStep))
}

func TestStepFit_SimpleStepUp(t *testing.T) {
	assert.Equal(t,
		&StepFit{TurningPoint: 2, StepSize: -2.0412414, Status: HIGH, Regression: -20.412415},
		GetStepFitAtMid([]float32{0, 0, 1, 1, 1}, minStdDev, 20, types.OriginalStep))
}

func TestStepFit_RealWorldExample(t *testing.T) {
	assert.Equal(t,
		&StepFit{LeastSquares: 0.17183419, TurningPoint: 1, StepSize: -2.0251877, Regression: -11.785709, Status: "High"},
		GetStepFitAtMid([]float32{608100,
			672970,
			653180}, minStdDev, 2, types.OriginalStep))
}

func TestStepFit_SimpleStepDown(t *testing.T) {
	assert.Equal(t,
		&StepFit{TurningPoint: 2, StepSize: 2.0412414, Status: LOW, Regression: 20.412415},
		GetStepFitAtMid([]float32{1, 1, 0, 0, 0}, minStdDev, 20, types.OriginalStep))
}

func TestStepFit_NoStep(t *testing.T) {
	assert.Equal(t,
		&StepFit{TurningPoint: 2, StepSize: -1, Status: UNINTERESTING, Regression: -2.7105057e-19,
			LeastSquares: 3.6893486e+18},
		GetStepFitAtMid([]float32{1, 1, 1, 1, 1}, minStdDev, 50, types.OriginalStep))
}

func TestStepFit_Absolute_NoStep(t *testing.T) {
	assert.Equal(t,
		&StepFit{TurningPoint: 2, StepSize: 0, Status: UNINTERESTING, Regression: 0, LeastSquares: InvalidLeastSquaresError},
		GetStepFitAtMid([]float32{1, 2, 1, 2, x}, minStdDev, 1.0, types.AbsoluteStep))
}

func TestStepFit_Absolute_StepExactMatch(t *testing.T) {
	assert.Equal(t,
		&StepFit{TurningPoint: 2, StepSize: -1, Status: HIGH, Regression: -1, LeastSquares: InvalidLeastSquaresError},
		GetStepFitAtMid([]float32{1, 1, 2, 2, x}, minStdDev, 1.0, types.AbsoluteStep))
}

func TestStepFit_Absolute_StepTooSmall(t *testing.T) {
	assert.Equal(t,
		&StepFit{TurningPoint: 2, StepSize: -0.5, Status: UNINTERESTING, Regression: -0.5, LeastSquares: InvalidLeastSquaresError},
		GetStepFitAtMid([]float32{1, 1, 1.5, 1.5, x}, minStdDev, 1.0, types.AbsoluteStep))
}

func TestStepFit_Const_NoRegression(t *testing.T) {
	assert.Equal(t,
		&StepFit{TurningPoint: 2, StepSize: -1, Status: UNINTERESTING, Regression: -1, LeastSquares: InvalidLeastSquaresError},
		GetStepFitAtMid([]float32{0, 0, 1, 0, x}, minStdDev, 2.0, types.Const))
}

func TestStepFit_Const_RegressionExactlyAtThreshhold(t *testing.T) {
	assert.Equal(t,
		&StepFit{TurningPoint: 2, StepSize: 0, Status: HIGH, Regression: -1, LeastSquares: InvalidLeastSquaresError},
		GetStepFitAtMid([]float32{0, 0, 1, 0, x}, minStdDev, 1.0, types.Const))
}

func TestStepFit_Const_RegressionAfterApplyingAbs(t *testing.T) {
	assert.Equal(t,
		&StepFit{TurningPoint: 2, StepSize: 1, Status: HIGH, Regression: -2, LeastSquares: InvalidLeastSquaresError},
		GetStepFitAtMid([]float32{0, 0, -2, 0, x}, minStdDev, 1.0, types.Const))
}

func TestStepFit_Percent_NoStep(t *testing.T) {
	assert.Equal(t,
		&StepFit{TurningPoint: 2, StepSize: 0, Status: UNINTERESTING, Regression: 0, LeastSquares: InvalidLeastSquaresError},
		GetStepFitAtMid([]float32{1, 2, 1, 2, x}, minStdDev, 1.0, types.PercentStep))
}

func TestStepFit_Percent_StepExactMatch(t *testing.T) {
	assert.Equal(t,
		&StepFit{TurningPoint: 1, StepSize: -1, Status: HIGH, Regression: -1, LeastSquares: InvalidLeastSquaresError},
		GetStepFitAtMid([]float32{1, 2, x}, minStdDev, 1.0, types.PercentStep))
}

func TestStepFit_Percent_StepTooSmall(t *testing.T) {
	assert.Equal(t,
		&StepFit{TurningPoint: 1, StepSize: -0.5, Status: UNINTERESTING, Regression: -0.5, LeastSquares: InvalidLeastSquaresError},
		GetStepFitAtMid([]float32{1, 1.5, x}, minStdDev, 1.0, types.PercentStep))
}

func TestStepFit_Percent_DefendAgainstNegativeInf(t *testing.T) {
	assert.Equal(t,
		&StepFit{TurningPoint: 1, StepSize: -math.MaxFloat32, Status: HIGH, Regression: -math.MaxFloat32, LeastSquares: InvalidLeastSquaresError},
		GetStepFitAtMid([]float32{0, 0, 1, 1}, minStdDev, 1.0, types.PercentStep))
}

func TestStepFit_Percent_DefendAgainstInf(t *testing.T) {
	assert.Equal(t,
		&StepFit{TurningPoint: 1, StepSize: math.MaxFloat32, Status: LOW, Regression: math.MaxFloat32, LeastSquares: InvalidLeastSquaresError},
		GetStepFitAtMid([]float32{0, 0, -1, -1}, minStdDev, 1.0, types.PercentStep))
}

func TestStepFit_Percent_DefendAgainstNaN(t *testing.T) {
	assert.Equal(t,
		&StepFit{TurningPoint: 1, StepSize: 0, Status: UNINTERESTING, Regression: 0, LeastSquares: InvalidLeastSquaresError},
		GetStepFitAtMid([]float32{0, 0, 0, 0}, minStdDev, 1.0, types.PercentStep))
}

func TestStepFit_Cohen_NoStep(t *testing.T) {
	assert.Equal(t,
		&StepFit{TurningPoint: 3, StepSize: -0.1999998, Status: UNINTERESTING, Regression: -0.1999998, LeastSquares: InvalidLeastSquaresError},
		GetStepFitAtMid([]float32{1, 1.1, 0.9, 1.02, 1.12, 0.92, x}, minStdDev, 1.0, types.CohenStep))
}

func TestStepFit_Cohen_Step(t *testing.T) {
	assert.Equal(t,
		&StepFit{TurningPoint: 3, StepSize: -0.1999998, Status: HIGH, Regression: -0.1999998, LeastSquares: InvalidLeastSquaresError},
		GetStepFitAtMid([]float32{1, 1.1, 0.9, 1.02, 1.12, 0.92, x}, minStdDev, 0.1, types.CohenStep))
}

func TestStepFit_Cohen_StepWithZeroStandardDeviation(t *testing.T) {
	assert.Equal(t,
		&StepFit{TurningPoint: 2, StepSize: -10, Status: HIGH, Regression: -10, LeastSquares: InvalidLeastSquaresError},
		GetStepFitAtMid([]float32{1, 1, 2, 2, x}, minStdDev, 0.2, types.CohenStep))
}
func TestStepFit_Cohen_StepWithLargeStandardDeviation(t *testing.T) {
	assert.Equal(t,
		&StepFit{LeastSquares: InvalidLeastSquaresError, TurningPoint: 2, StepSize: -2.828427, Regression: -2.828427, Status: "High"},
		GetStepFitAtMid([]float32{1, 2, 3, 4, x}, minStdDev, 0.2, types.CohenStep))
}

func TestStepFit_MannWhitneyU_StepHigh(t *testing.T) {
	assert.Equal(t,
		&StepFit{LeastSquares: 0, TurningPoint: 4, StepSize: -10, Regression: -0.028571428571428577 /* value copied from stats/utest_test.go:49 */, Status: "High"},
		GetStepFitAtMid([]float32{2, 1, 3, 5, 12, 11, 13, 15, x}, minStdDev, 0.05, types.MannWhitneyU))
}

func TestStepFit_MannWhitneyU_StepLow(t *testing.T) {
	assert.Equal(t,
		&StepFit{LeastSquares: 16, TurningPoint: 4, StepSize: 10, Regression: 0.028571428571428577 /* value copied from stats/utest_test.go:49 */, Status: "Low"},
		GetStepFitAtMid([]float32{12, 11, 13, 15, 2, 1, 3, 5, x}, minStdDev, 0.05, types.MannWhitneyU))
}

func TestStepFit_MannWhitneyU_UninterestingBecauseInterestingThreshholdTooLow(t *testing.T) {
	assert.Equal(t,
		&StepFit{LeastSquares: 16, TurningPoint: 4, StepSize: 10, Regression: 0.028571428571428577 /* value copied from stats/utest_test.go:49 */, Status: "Uninteresting"},
		GetStepFitAtMid([]float32{12, 11, 13, 15, 2, 1, 3, 5, x}, minStdDev, 0.01, types.MannWhitneyU))
}

func TestStepFit_MannWhitneyU_UninterestingBecauseBothSidesAreEqual(t *testing.T) {
	assert.Equal(t,
		&StepFit{LeastSquares: 8, TurningPoint: 4, StepSize: 0, Regression: 1, Status: "Uninteresting"},
		GetStepFitAtMid([]float32{2, 1, 3, 5, 2, 1, 3, 5, x}, minStdDev, 0.01, types.MannWhitneyU))
}

func TestStepFit_MannWhitneyU_UninterestingBecauseBothSidesAreConstant(t *testing.T) {
	assert.Equal(t,
		&StepFit{LeastSquares: 0, TurningPoint: 0, StepSize: 0, Regression: 0, Status: "Uninteresting"},
		GetStepFitAtMid([]float32{1, 1, 1, 1, 1, 1, 1, 1, x}, minStdDev, 0.01, types.MannWhitneyU))
}

func TestStepFit_MannWhitneyU_UninterestingBecauseNotEnoughData(t *testing.T) {
	assert.Equal(t,
		&StepFit{LeastSquares: 0, TurningPoint: 0, StepSize: 0, Regression: 0, Status: "Uninteresting"},
		GetStepFitAtMid([]float32{2, 2, x}, minStdDev, 0.01, types.MannWhitneyU))
}
