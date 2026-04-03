package validate

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInsertPending_InsertsWithCorrectStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	event := &ValidationEvent{
		ID:        "evt-123",
		RunID:     "run-456",
		Name:      "test-validation",
		Script:    "echo test",
		Status:    StatusPending,
		CreatedAt: time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC),
	}

	mock.ExpectExec(regexp.QuoteMeta(insertPendingQuery)).WithArgs(
		"evt-123", "run-456", "test-validation", "echo test", event.CreatedAt,
	).WillReturnResult(sqlmock.NewResult(1, 1))

	err = InsertPending(db, event)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRecordResult_UpdatesStatusAndResult(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(updateResultQuery)).WithArgs(
		ResultPass, sqlmock.AnyArg(), "test-validation", "run-456",
	).WillReturnResult(sqlmock.NewResult(0, 1))

	err = RecordResult(db, "test-validation", "run-456", ResultPass)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRecordResult_WithDifferentResults(t *testing.T) {
	tests := []struct {
		name   string
		result string
	}{
		{"pass result", ResultPass},
		{"fail result", ResultFail},
		{"irrelevant result", ResultIrrelevant},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, mock, err := sqlmock.New()
			require.NoError(t, err)
			defer db.Close()

			mock.ExpectExec(regexp.QuoteMeta(updateResultQuery)).WithArgs(
				tt.result, sqlmock.AnyArg(), "validation-name", "run-123",
			).WillReturnResult(sqlmock.NewResult(0, 1))

			err = RecordResult(db, "validation-name", "run-123", tt.result)
			require.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRecordResult_ReturnsErrorWhenNoPendingRecord(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(updateResultQuery)).WithArgs(
		ResultPass, sqlmock.AnyArg(), "nonexistent-validation", "run-789",
	).WillReturnResult(sqlmock.NewResult(0, 0))

	err = RecordResult(db, "nonexistent-validation", "run-789", ResultPass)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no pending validation found")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRecordResult_ReturnsErrorOnDBFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(updateResultQuery)).WithArgs(
		ResultPass, sqlmock.AnyArg(), "test-validation", "run-456",
	).WillReturnError(assert.AnError)

	err = RecordResult(db, "test-validation", "run-456", ResultPass)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMarkOrphaned_MarksPendingFromDifferentRunIDs(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	currentRunID := "run-current"

	mock.ExpectExec(regexp.QuoteMeta(markOrphanedQuery)).WithArgs(
		sqlmock.AnyArg(), currentRunID,
	).WillReturnResult(sqlmock.NewResult(0, 3))

	err = MarkOrphaned(db, currentRunID)
	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMarkOrphaned_DoesNotMarkCurrentRunID(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	currentRunID := "run-abc-123"

	mock.ExpectExec(regexp.QuoteMeta(markOrphanedQuery)).WithArgs(
		sqlmock.AnyArg(), currentRunID,
	).WillReturnResult(sqlmock.NewResult(0, 0))

	err = MarkOrphaned(db, currentRunID)
	require.NoError(t, err)

	result := sqlmock.NewResult(0, 0)
	rowsAffected, _ := result.RowsAffected()
	assert.Equal(t, int64(0), rowsAffected, "No records should be marked orphaned for current run")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMarkOrphaned_ReturnsErrorOnDBFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec(regexp.QuoteMeta(markOrphanedQuery)).WithArgs(
		sqlmock.AnyArg(), "run-123",
	).WillReturnError(assert.AnError)

	err = MarkOrphaned(db, "run-123")
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestInsertPending_ReturnsErrorOnDBFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	event := &ValidationEvent{
		ID:        "evt-123",
		RunID:     "run-456",
		Name:      "test-validation",
		Script:    "echo test",
		CreatedAt: time.Now(),
	}

	mock.ExpectExec(regexp.QuoteMeta(insertPendingQuery)).WithArgs(
		event.ID, event.RunID, event.Name, event.Script, event.CreatedAt,
	).WillReturnError(assert.AnError)

	err = InsertPending(db, event)
	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
