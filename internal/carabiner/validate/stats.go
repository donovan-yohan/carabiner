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
	query := validationStatsQuery

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
