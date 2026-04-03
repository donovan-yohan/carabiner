package validate

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidationStats_EmptyDatabase(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	query := regexp.QuoteMeta(`SELECT name, SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending, SUM(CASE WHEN status = 'responded' THEN 1 ELSE 0 END) as responded, SUM(CASE WHEN status = 'orphaned' THEN 1 ELSE 0 END) as orphaned, MAX(created_at) as last_run FROM validation_events GROUP BY name ORDER BY name`)

	mock.ExpectQuery(query).WillReturnRows(sqlmock.NewRows([]string{"name", "pending", "responded", "orphaned", "last_run"}))

	stats, err := ValidationStats(db)
	require.NoError(t, err)
	assert.Empty(t, stats)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestValidationStats_SingleValidation(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	query := regexp.QuoteMeta(`SELECT name, SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending, SUM(CASE WHEN status = 'responded' THEN 1 ELSE 0 END) as responded, SUM(CASE WHEN status = 'orphaned' THEN 1 ELSE 0 END) as orphaned, MAX(created_at) as last_run FROM validation_events GROUP BY name ORDER BY name`)

	lastRun := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	rows := sqlmock.NewRows([]string{"name", "pending", "responded", "orphaned", "last_run"}).
		AddRow("auth-check", 2, 5, 1, lastRun)

	mock.ExpectQuery(query).WillReturnRows(rows)

	stats, err := ValidationStats(db)
	require.NoError(t, err)
	require.Len(t, stats, 1)

	stat := stats[0]
	assert.Equal(t, "auth-check", stat.Name)
	assert.Equal(t, 2, stat.Pending)
	assert.Equal(t, 5, stat.Responded)
	assert.Equal(t, 1, stat.Orphaned)
	assert.NotNil(t, stat.LastRun)
	assert.Equal(t, lastRun, *stat.LastRun)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestValidationStats_MultipleValidations(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	query := regexp.QuoteMeta(`SELECT name, SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending, SUM(CASE WHEN status = 'responded' THEN 1 ELSE 0 END) as responded, SUM(CASE WHEN status = 'orphaned' THEN 1 ELSE 0 END) as orphaned, MAX(created_at) as last_run FROM validation_events GROUP BY name ORDER BY name`)

	lastRun1 := time.Date(2024, 1, 10, 9, 0, 0, 0, time.UTC)
	lastRun2 := time.Date(2024, 1, 15, 14, 30, 0, 0, time.UTC)

	rows := sqlmock.NewRows([]string{"name", "pending", "responded", "orphaned", "last_run"}).
		AddRow("auth-check", 1, 3, 0, lastRun1).
		AddRow("lint-check", 0, 10, 2, lastRun2)

	mock.ExpectQuery(query).WillReturnRows(rows)

	stats, err := ValidationStats(db)
	require.NoError(t, err)
	require.Len(t, stats, 2)

	stat1 := stats[0]
	assert.Equal(t, "auth-check", stat1.Name)
	assert.Equal(t, 1, stat1.Pending)
	assert.Equal(t, 3, stat1.Responded)
	assert.Equal(t, 0, stat1.Orphaned)
	assert.Equal(t, lastRun1, *stat1.LastRun)

	stat2 := stats[1]
	assert.Equal(t, "lint-check", stat2.Name)
	assert.Equal(t, 0, stat2.Pending)
	assert.Equal(t, 10, stat2.Responded)
	assert.Equal(t, 2, stat2.Orphaned)
	assert.Equal(t, lastRun2, *stat2.LastRun)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestValidationStats_WithNullLastRun(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	query := regexp.QuoteMeta(`SELECT name, SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending, SUM(CASE WHEN status = 'responded' THEN 1 ELSE 0 END) as responded, SUM(CASE WHEN status = 'orphaned' THEN 1 ELSE 0 END) as orphaned, MAX(created_at) as last_run FROM validation_events GROUP BY name ORDER BY name`)

	rows := sqlmock.NewRows([]string{"name", "pending", "responded", "orphaned", "last_run"}).
		AddRow("new-validation", 1, 0, 0, nil)

	mock.ExpectQuery(query).WillReturnRows(rows)

	stats, err := ValidationStats(db)
	require.NoError(t, err)
	require.Len(t, stats, 1)

	stat := stats[0]
	assert.Equal(t, "new-validation", stat.Name)
	assert.Equal(t, 1, stat.Pending)
	assert.Nil(t, stat.LastRun)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestValidationStats_ReturnsErrorOnQueryFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	query := regexp.QuoteMeta(`SELECT name, SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending, SUM(CASE WHEN status = 'responded' THEN 1 ELSE 0 END) as responded, SUM(CASE WHEN status = 'orphaned' THEN 1 ELSE 0 END) as orphaned, MAX(created_at) as last_run FROM validation_events GROUP BY name ORDER BY name`)

	mock.ExpectQuery(query).WillReturnError(assert.AnError)

	stats, err := ValidationStats(db)
	assert.Error(t, err)
	assert.Nil(t, stats)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestValidationStats_ReturnsErrorOnScanFailure(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	query := regexp.QuoteMeta(`SELECT name, SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending, SUM(CASE WHEN status = 'responded' THEN 1 ELSE 0 END) as responded, SUM(CASE WHEN status = 'orphaned' THEN 1 ELSE 0 END) as orphaned, MAX(created_at) as last_run FROM validation_events GROUP BY name ORDER BY name`)

	rows := sqlmock.NewRows([]string{"name", "pending", "responded", "orphaned", "last_run"}).
		AddRow("test", "invalid", 0, 0, time.Now())

	mock.ExpectQuery(query).WillReturnRows(rows)

	stats, err := ValidationStats(db)
	assert.Error(t, err)
	assert.Nil(t, stats)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestValidationStats_VariousStatuses(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	query := regexp.QuoteMeta(`SELECT name, SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending, SUM(CASE WHEN status = 'responded' THEN 1 ELSE 0 END) as responded, SUM(CASE WHEN status = 'orphaned' THEN 1 ELSE 0 END) as orphaned, MAX(created_at) as last_run FROM validation_events GROUP BY name ORDER BY name`)

	lastRun := time.Now()

	rows := sqlmock.NewRows([]string{"name", "pending", "responded", "orphaned", "last_run"}).
		AddRow("all-pending", 5, 0, 0, lastRun).
		AddRow("all-responded", 0, 5, 0, lastRun).
		AddRow("all-orphaned", 0, 0, 5, lastRun).
		AddRow("mixed", 2, 3, 1, lastRun)

	mock.ExpectQuery(query).WillReturnRows(rows)

	stats, err := ValidationStats(db)
	require.NoError(t, err)
	require.Len(t, stats, 4)

	assert.Equal(t, "all-pending", stats[0].Name)
	assert.Equal(t, 5, stats[0].Pending)
	assert.Equal(t, 0, stats[0].Responded)
	assert.Equal(t, 0, stats[0].Orphaned)

	assert.Equal(t, "all-responded", stats[1].Name)
	assert.Equal(t, 0, stats[1].Pending)
	assert.Equal(t, 5, stats[1].Responded)
	assert.Equal(t, 0, stats[1].Orphaned)

	assert.Equal(t, "all-orphaned", stats[2].Name)
	assert.Equal(t, 0, stats[2].Pending)
	assert.Equal(t, 0, stats[2].Responded)
	assert.Equal(t, 5, stats[2].Orphaned)

	assert.Equal(t, "mixed", stats[3].Name)
	assert.Equal(t, 2, stats[3].Pending)
	assert.Equal(t, 3, stats[3].Responded)
	assert.Equal(t, 1, stats[3].Orphaned)

	assert.NoError(t, mock.ExpectationsWereMet())
}
