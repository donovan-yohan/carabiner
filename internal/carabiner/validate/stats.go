package validate

import (
	"database/sql"
	"time"
)

type ValidationStat struct {
	Name      string
	Pending   int
	Responded int
	Orphaned  int
	LastRun   *time.Time
}

type StatsByName []ValidationStat

// ValidationStats returns aggregated stats per validation name.
func ValidationStats(db *sql.DB) (StatsByName, error) {
	query := `
	SELECT 
		name,
		SUM(CASE WHEN status = 'pending' THEN 1 ELSE 0 END) as pending,
		SUM(CASE WHEN status = 'responded' THEN 1 ELSE 0 END) as responded,
		SUM(CASE WHEN status = 'orphaned' THEN 1 ELSE 0 END) as orphaned,
		MAX(created_at) as last_run
	FROM validation_events
	GROUP BY name
	ORDER BY name`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats StatsByName
	for rows.Next() {
		var s ValidationStat
		var lastRun sql.NullTime
		err := rows.Scan(&s.Name, &s.Pending, &s.Responded, &s.Orphaned, &lastRun)
		if err != nil {
			return nil, err
		}
		if lastRun.Valid {
			s.LastRun = &lastRun.Time
		}
		stats = append(stats, s)
	}
	return stats, rows.Err()
}
