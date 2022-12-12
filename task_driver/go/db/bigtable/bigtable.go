package bigtable

/*
   This DB implementation uses BigTable to store information about Task Drivers.
*/

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"time"

	"cloud.google.com/go/bigtable"
	"go.opencensus.io/trace"
	"golang.org/x/oauth2"
	"google.golang.org/api/option"

	"go.skia.org/infra/go/util"
	"go.skia.org/infra/task_driver/go/db"
	"go.skia.org/infra/task_driver/go/td"
)

const (
	// We use a single BigTable table for storing Task Driver runs.
	btTable = "task-driver-runs"

	// We use a single BigTable column family.
	btColumnFamily = "MSGS"

	// We use a single BigTable column which stores gob-encoded td.Messages.
	btColumn = "MSG"

	// Format used for BigTable row keys.
	rowKeyFormat           = "%s#%s#%s"
	rowKeyFormatDeprecated = "%s#%010d"

	insertTimeout = 30 * time.Second
	queryTimeout  = 5 * time.Second
)

var (
	// Fully-qualified BigTable column name.
	btColumnFull = fmt.Sprintf("%s:%s", btColumnFamily, btColumn)
)

// rowKey returns a BigTable row key for the given message, based on the given
// Task Driver ID.
func rowKey(id string, msg *td.Message) string {
	if msg.ID == "" {
		return fmt.Sprintf(rowKeyFormatDeprecated, id, msg.Index)
	}
	return fmt.Sprintf(rowKeyFormat, id, msg.Timestamp.Format(util.SAFE_TIMESTAMP_FORMAT), msg.ID)
}

// BTDB is an implementation of db.DB which uses BigTable.
type BTDB struct {
	client *bigtable.Client
	table  *bigtable.Table
}

// NewBigTableDB returns a db.DB instance which uses BigTable.
func NewBigTableDB(ctx context.Context, project, instance string, ts oauth2.TokenSource) (*BTDB, error) {
	client, err := bigtable.NewClient(ctx, project, instance, option.WithTokenSource(ts))
	if err != nil {
		return nil, fmt.Errorf("Failed to create BigTable client: %s", err)
	}
	table := client.Open(btTable)
	return &BTDB{
		client: client,
		table:  table,
	}, nil
}

// See documentation for db.DB interface.
func (d *BTDB) Close() error {
	return d.client.Close()
}

// GetMessagesForTaskDriver returns all td.Messages sent for the Task Driver
// with the given ID.
func (d *BTDB) GetMessagesForTaskDriver(ctx context.Context, id string) ([]*td.Message, error) {
	ctx, span := trace.StartSpan(ctx, "bigtable_GetMessagesForTaskDriver")
	defer span.End()
	// Retrieve all messages for the Task Driver from BigTable.
	msgs := []*td.Message{}
	var decodeErr error
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()
	if err := d.table.ReadRows(ctx, bigtable.PrefixRange(id), func(row bigtable.Row) bool {
		for _, ri := range row[btColumnFamily] {
			if ri.Column == btColumnFull {
				var msg td.Message
				decodeErr = gob.NewDecoder(bytes.NewReader(ri.Value)).Decode(&msg)
				if decodeErr != nil {
					return false
				}
				msgs = append(msgs, &msg)
				// We only store one message per row.
				return true
			}
		}
		return true
	}, bigtable.RowFilter(bigtable.LatestNFilter(1))); err != nil {
		return nil, fmt.Errorf("Failed to retrieve data from BigTable: %s", err)
	}
	if decodeErr != nil {
		return nil, fmt.Errorf("Failed to gob-decode message: %s", decodeErr)
	}
	return msgs, nil
}

// See documentation for db.DB interface.
func (d *BTDB) GetTaskDriver(ctx context.Context, id string) (*db.TaskDriverRun, error) {
	ctx, span := trace.StartSpan(ctx, "bigtable_GetTaskDriver")
	defer span.End()
	msgs, err := d.GetMessagesForTaskDriver(ctx, id)
	if err != nil {
		return nil, err
	}

	// If we have no messages, the TaskDriverRun does not exist in our DB.
	// Per the doc on the db.DB interface, we should return nil with no
	// error.
	if len(msgs) == 0 {
		return nil, nil
	}

	// Apply all messages to the TaskDriver.
	t := &db.TaskDriverRun{
		TaskId: id,
	}
	for _, msg := range msgs {
		if err := t.UpdateFromMessage(msg); err != nil {
			return nil, fmt.Errorf("Failed to apply update to TaskDriverRun: %s", err)
		}
	}
	return t, nil
}

// See documentation for db.DB interface.
func (d *BTDB) UpdateTaskDriver(ctx context.Context, id string, msg *td.Message) error {
	ctx, span := trace.StartSpan(ctx, "bigtable_UpdateTaskDriver")
	defer span.End()
	// Encode the Message.
	buf := bytes.Buffer{}
	if err := gob.NewEncoder(&buf).Encode(msg); err != nil {
		return fmt.Errorf("Failed to gob-encode Message: %s", err)
	}
	// Insert the message into BigTable.
	mt := bigtable.NewMutation()
	mt.Set(btColumnFamily, btColumn, bigtable.Time(msg.Timestamp), buf.Bytes())
	rk := rowKey(id, msg)
	ctx, cancel := context.WithTimeout(ctx, insertTimeout)
	defer cancel()
	return d.table.Apply(ctx, rk, mt)
}

var _ db.DB = &BTDB{}
