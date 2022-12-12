package firestore

import (
	"testing"

	"go.skia.org/infra/task_scheduler/go/db"
)

func TestModifiedTasksCh(t *testing.T) {
	d, cleanup := setup(t)
	defer cleanup()
	db.TestModifiedTasksCh(t, d)
}

func TestModifiedJobsCh(t *testing.T) {
	d, cleanup := setup(t)
	defer cleanup()
	db.TestModifiedJobsCh(t, d)
}

func TestModifiedTaskCommentsCh(t *testing.T) {
	d, cleanup := setup(t)
	defer cleanup()
	db.TestModifiedTaskCommentsCh(t, d)
}

func TestModifiedTaskSpecCommentsCh(t *testing.T) {
	d, cleanup := setup(t)
	defer cleanup()
	db.TestModifiedTaskSpecCommentsCh(t, d)
}

func TestModifiedCommitCommentsCh(t *testing.T) {
	d, cleanup := setup(t)
	defer cleanup()
	db.TestModifiedCommitCommentsCh(t, d)
}
