package status

import (
	"bytes"
	"context"
	"encoding/gob"
	"time"

	"cloud.google.com/go/datastore"
	"go.skia.org/infra/autoroll/go/revision"
	"go.skia.org/infra/go/autoroll"
	"go.skia.org/infra/go/ds"
	"go.skia.org/infra/go/sklog"
	"go.skia.org/infra/go/util"
)

var (
	// Insert the AutoRollMiniStatus for these internal rollers into the
	// external datastore.
	exportRollers = []string{
		"android-master-autoroll",
		"google3-autoroll",
	}

	// dsInitialized lets us know whether we've initialized Datastore for the
	// autoroller status DB.
	dsInitialized = false
)

// AutoRollStatus is a struct which provides roll-up status information about
// the AutoRoll Bot.
type AutoRollStatus struct {
	AutoRollMiniStatus
	ChildHead          string                    `json:"childHead"`
	ChildName          string                    `json:"childName"`
	CurrentRoll        *autoroll.AutoRollIssue   `json:"currentRoll"`
	Error              string                    `json:"error"`
	FullHistoryUrl     string                    `json:"fullHistoryUrl"`
	IssueUrlBase       string                    `json:"issueUrlBase"`
	LastRoll           *autoroll.AutoRollIssue   `json:"lastRoll"`
	NotRolledRevisions []*revision.Revision      `json:"notRolledRevs"`
	ParentName         string                    `json:"parentName"`
	Recent             []*autoroll.AutoRollIssue `json:"recent"`
	Status             string                    `json:"status"`
	ThrottledUntil     int64                     `json:"throttledUntil"`
	ValidModes         []string                  `json:"validModes"`
	ValidStrategies    []string                  `json:"validStrategies"`
}

func (s *AutoRollStatus) Copy() *AutoRollStatus {
	if s == nil {
		sklog.Warningf("Copying nil AutoRollStatus.")
		return nil
	}
	var recent []*autoroll.AutoRollIssue
	if s.Recent != nil {
		recent = make([]*autoroll.AutoRollIssue, 0, len(s.Recent))
		for _, r := range s.Recent {
			recent = append(recent, r.Copy())
		}
	}
	var notRolledRevisions []*revision.Revision
	if s.NotRolledRevisions != nil {
		notRolledRevisions = make([]*revision.Revision, 0, len(s.NotRolledRevisions))
		for _, r := range s.NotRolledRevisions {
			notRolledRevisions = append(notRolledRevisions, r.Copy())
		}
	}
	rv := &AutoRollStatus{
		AutoRollMiniStatus: AutoRollMiniStatus{
			CurrentRollRev:      s.CurrentRollRev,
			LastRollRev:         s.LastRollRev,
			NumFailedRolls:      s.NumFailedRolls,
			NumNotRolledCommits: s.NumNotRolledCommits,
			Timestamp:           s.Timestamp,
		},
		ChildHead:          s.ChildHead,
		ChildName:          s.ChildName,
		Error:              s.Error,
		FullHistoryUrl:     s.FullHistoryUrl,
		IssueUrlBase:       s.IssueUrlBase,
		NotRolledRevisions: notRolledRevisions,
		ParentName:         s.ParentName,
		Recent:             recent,
		Status:             s.Status,
		ThrottledUntil:     s.ThrottledUntil,
		ValidModes:         util.CopyStringSlice(s.ValidModes),
		ValidStrategies:    util.CopyStringSlice(s.ValidStrategies),
	}
	if s.CurrentRoll != nil {
		rv.CurrentRoll = s.CurrentRoll.Copy()
	}
	if s.LastRoll != nil {
		rv.LastRoll = s.LastRoll.Copy()
	}
	return rv
}

// AutoRollMiniStatus is a struct which provides a minimal amount of status
// information about the AutoRoll Bot.
// TODO(borenet): Some of this duplicates things in AutoRollStatus. Revisit and
// either don't include AutoRollMiniStatus in AutoRollStatus or de-dupe the
// fields after revamping the UI.
type AutoRollMiniStatus struct {
	// Revision of the current roll, if any.
	CurrentRollRev string `json:"currentRollRev"`

	// Revision of the last successful roll.
	LastRollRev string `json:"lastRollRev"`

	// The current mode of the roller.
	// Note: This duplicates what is stored in the modes DB but is more
	// convenient for users like status.skia.org.
	Mode string `json:"mode"`

	// The number of failed rolls since the last successful roll.
	NumFailedRolls int `json:"numFailed"`

	// The number of commits which have not been rolled.
	NumNotRolledCommits int `json:"numBehind"`

	// Timestamp is the time at which the roller last reported its status.
	Timestamp time.Time `json:"timestamp"`
}

// DB tracks the status of the autoroller.
type DB interface {
	// Get the most recent AutoRollStatus for the given roller.
	Get(ctx context.Context, rollerName string) (*AutoRollStatus, error)
	// Set a new AutoRollStatus for the given roller.
	Set(ctx context.Context, rollerName string, st *AutoRollStatus) error
}

// Fake ancestor we supply for all AutoRollStatus, to force strong consistency.
// We lose some performance this way but it keeps our tests from flaking.
func fakeAncestor() *datastore.Key {
	rv := ds.NewKey(ds.KIND_AUTOROLL_STATUS_ANCESTOR)
	rv.ID = 13 // Bogus ID.
	return rv
}

// DsStatusWrapper is a helper struct used for storing an AutoRollStatus in the
// datastore.
type DsStatusWrapper struct {
	Data   []byte `datastore:"data,noindex"`
	Roller string `datastore:"roller"`
}

// Create a key for the given roller.
func key(rollerName string) *datastore.Key {
	key := ds.NewKey(ds.KIND_AUTOROLL_STATUS)
	key.Name = rollerName
	key.Parent = fakeAncestor()
	return key
}

// DatastoreDB is a DB which uses Datastore.
type DatastoreDB struct{}

// NewDatastoreDB returns a DatastoreDB instance.
func NewDatastoreDB() *DatastoreDB {
	return &DatastoreDB{}
}

// Set the AutoRollStatus for the given roller in the datastore.
func (d *DatastoreDB) Set(ctx context.Context, rollerName string, st *AutoRollStatus) error {
	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(st); err != nil {
		return err
	}
	w := &DsStatusWrapper{
		Data:   buf.Bytes(),
		Roller: rollerName,
	}
	_, err := ds.DS.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		_, err := tx.Put(key(rollerName), w)
		return err
	})
	if err != nil {
		return err
	}

	// Optionally export the mini version of the internal roller's status
	// to the external datastore.
	if util.In(rollerName, exportRollers) {
		exportStatus := &AutoRollStatus{
			AutoRollMiniStatus: st.AutoRollMiniStatus,
		}
		buf := bytes.NewBuffer(nil)
		if err := gob.NewEncoder(buf).Encode(exportStatus); err != nil {
			return err
		}
		w := &DsStatusWrapper{
			Data:   buf.Bytes(),
			Roller: rollerName,
		}
		_, err := ds.DS.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
			k := key(rollerName)
			k.Namespace = ds.AUTOROLL_NS
			k.Parent.Namespace = ds.AUTOROLL_NS
			_, err := tx.Put(k, w)
			return err
		})
		if err != nil {
			return err
		}
	}
	return nil
}

// Get the AutoRollStatus for the given roller from the datastore.
func (d *DatastoreDB) Get(ctx context.Context, rollerName string) (*AutoRollStatus, error) {
	var w DsStatusWrapper
	if err := ds.DS.Get(ctx, key(rollerName), &w); err != nil {
		return nil, err
	}
	rv := new(AutoRollStatus)
	if err := gob.NewDecoder(bytes.NewReader(w.Data)).Decode(rv); err != nil {
		return nil, err
	}
	return rv, nil
}

var _ DB = &DatastoreDB{}
