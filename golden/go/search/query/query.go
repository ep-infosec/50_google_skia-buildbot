// Package query contains the logic involving parsing queries to
// Gold's search endpoints.
package query

import (
	"net/http"

	"go.skia.org/infra/go/skerr"
	"go.skia.org/infra/go/util"
	"go.skia.org/infra/golden/go/types"
	"go.skia.org/infra/golden/go/validation"
)

const (
	// SortAscending indicates that we want to sort in ascending order.
	SortAscending = "asc"
	// SortDescending indicates that we want to sort in descending order.
	SortDescending = "desc"

	// CombinedMetric corresponds to diff.DiffMetric.CombinedMetric
	CombinedMetric = "combined"
	// PercentMetric corresponds to diff.DiffMetric.PixelDiffPercent
	PercentMetric = "percent"
	// PixelMetric corresponds to diff.DiffMetric.NumDiffPixels
	PixelMetric = "pixel"
)

// ParseSearch parses the request parameters from the URL query string or from the
// form parameters and stores the parsed and validated values in query.
func ParseSearch(r *http.Request, q *Search) error {
	if err := r.ParseForm(); err != nil {
		return skerr.Wrapf(err, "parsing form")
	}

	// Parse the list of fields that need to match and ensure the
	// test name is in it.
	var ok bool
	if q.Match, ok = r.Form["match"]; ok {
		if !util.In(types.PrimaryKeyField, q.Match) {
			q.Match = append(q.Match, types.PrimaryKeyField)
		}
	} else {
		q.Match = []string{types.PrimaryKeyField}
	}

	validate := validation.Validation{}

	// Parse the query strings.
	q.TraceValues = validate.QueryFormValue(r, "query")
	q.RightTraceValues = validate.QueryFormValue(r, "rquery")

	q.Limit = int(validate.Int64FormValue(r, "limit", 50))
	q.Offset = int(validate.Int64FormValue(r, "offset", 0))
	q.Offset = util.MaxInt(q.Offset, 0)

	validate.StrFormValue(r, "metric", &q.Metric, []string{CombinedMetric, PercentMetric, PixelMetric}, CombinedMetric)
	validate.StrFormValue(r, "sort", &q.Sort, []string{SortDescending, SortAscending}, SortDescending)

	// Parse and validate the filter values.
	q.RGBAMinFilter = int(validate.Int64FormValue(r, "frgbamin", 0))
	q.RGBAMaxFilter = int(validate.Int64FormValue(r, "frgbamax", 255))

	// Parse out the issue and patchsets.
	q.Patchsets = validate.Int64SliceFormValue(r, "patchsets", nil)
	q.ChangelistID = r.FormValue("issue")
	q.CodeReviewSystemID = r.FormValue("crs")

	// Check whether any of the validations failed.
	if err := validate.Errors(); err != nil {
		return skerr.Wrapf(err, "validating params")
	}

	q.BlameGroupID = r.FormValue("blame")
	q.IncludePositiveDigests = r.FormValue("pos") == "true"
	q.IncludeNegativeDigests = r.FormValue("neg") == "true"
	q.IncludeUntriagedDigests = r.FormValue("unt") == "true"
	q.OnlyIncludeDigestsProducedAtHead = r.FormValue("head") == "true"
	q.IncludeIgnoredTraces = r.FormValue("include") == "true"
	// TODO(kjlubick) rename this
	q.IncludeDigestsProducedOnMaster = r.FormValue("master") == "true"

	// Extract the filter values.
	q.MustIncludeReferenceFilter = r.FormValue("fref") == "true"

	return nil
}
