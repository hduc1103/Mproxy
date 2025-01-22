package handlers

import (
	"database/sql"
	"log"
	"time"
)

func CheckSpam(db *sql.DB, deviceID string) (bool, error) {
	query := `SELECT timestamp FROM messages WHERE device_id = ? ORDER BY timestamp DESC LIMIT 5`

	rows, err := db.Query(query, deviceID)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("No previous messages for device %s. Not spamming.\n", deviceID)
			return false, nil
		}
		log.Printf("Error querying timestamps for device %s: %v\n", deviceID, err)
		return false, err
	}
	defer rows.Close()

	var timestamps []time.Time
	for rows.Next() {
		var timestamp time.Time
		if err := rows.Scan(&timestamp); err != nil {
			log.Printf("Failed to scan row: %v", err)
			return false, err
		}
		timestamps = append(timestamps, timestamp)
	}

	if len(timestamps) < 5 {
		return false, nil
	}

	cur_time := time.Now()
	if cur_time.Sub(timestamps[0]) > 30*time.Minute {
		return false, nil
	}

	var total_gap time.Duration
	for i := 0; i < len(timestamps)-1; i++ {
		total_gap += timestamps[i].Sub(timestamps[i+1])
	}

	average_gap := total_gap / time.Duration(len(timestamps)-1)
	is_spamming := average_gap < 1*time.Minute

	if is_spamming {
		log.Printf("Device %s is spamming. Average gap: %v\n", deviceID, average_gap)
	} else {
		log.Printf("Device %s is not spamming. Average gap: %v\n", deviceID, average_gap)
	}

	return is_spamming, nil
}
