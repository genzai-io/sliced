package btrdb

import (
	"errors"
	"fmt"
	"strings"
)

// Special file path that will not persist to disk.
const InMemoryPath = ":memory:"

var (
	// ErrTxNotWritable is returned when performing a write operation on a
	// read-only transaction.
	ErrTxNotWritable = errors.New("tx not writable")

	// ErrTxClosed is returned when committing or rolling back a transaction
	// that has already been committed or rolled back.
	ErrTxClosed = errors.New("tx closed")

	// ErrNotFound is returned when an item or index is not in the database.
	ErrNotFound = errors.New("not found")

	// ErrInvalid is returned when the database file is an invalid format.
	ErrInvalid = errors.New("invalid database")

	// ErrDatabaseClosed is returned when the database is closed.
	ErrDatabaseClosed = errors.New("database closed")

	// ErrIndexExists is returned when an index already exists in the database.
	ErrIndexExists = errors.New("index exists")

	// ErrInvalidOperation is returned when an operation cannot be completed.
	ErrInvalidOperation = errors.New("invalid operation")

	// ErrInvalidSyncPolicy is returned for an invalid SyncPolicy value.
	ErrInvalidSyncPolicy = errors.New("invalid sync policy")

	// ErrShrinkInProcess is returned when a shrink operation is in-process.
	ErrShrinkInProcess = errors.New("shrink is in-process")

	// ErrPersistenceActive is returned when post-loading data from an database
	// not opened with Open(":memory:").
	ErrPersistenceActive = errors.New("persistence active")

	// ErrTxIterating is returned when Set or Delete are called while iterating.
	ErrTxIterating = errors.New("tx is iterating")

	// ErrDuplicateKey is returned when Set is called on an existing key
	ErrDuplicateKey = errors.New("duplicate key")

	ErrUnexpectedDocument = errors.New("unexpected document")
	ErrNilDocument        = errors.New("nil document")

	ErrNeedsInit      = errors.New("needs init")
	ErrPKIndexMissing = errors.New("pk index missing")
	ErrPkMissing      = errors.New("pk missing")
	ErrPKNotNumber    = errors.New("pk not a number")

	ErrEmptySchema = errors.New("empty schema")
	ErrSealed      = errors.New("sealed")
	ErrMissingKey  = errors.New("missing key")
)

func ErrMarshaling(err error) error {
	return fmt.Errorf("marshalling: %s", err.Error())
}

func IsErrMarshaling(err error) bool {
	return strings.Index(err.Error(), "marshalling:") == 0
}

func ErrUnMarshaling(err error) error {
	return fmt.Errorf("unmarshalling: %s", err.Error())
}

func IsErrUnMarshaling(err error) bool {
	return strings.Index(err.Error(), "unmarshalling:") == 0
}
