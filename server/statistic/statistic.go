package main

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"github.com/wcharczuk/go-chart/v2"
)

func fetchMessageCounts(db *sql.DB) ([]string, []float64) {
	query := "SELECT device_id, COUNT(*) AS message_count FROM messages GROUP BY device_id ORDER BY message_count DESC"
	rows, err := db.Query(query)
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}
	defer rows.Close()

	var deviceIDs []string
	var messageCounts []float64
	for rows.Next() {
		var deviceID string
		var messageCount int
		if err := rows.Scan(&deviceID, &messageCount); err != nil {
			log.Fatalf("Failed to scan row: %v", err)
		}
		deviceIDs = append(deviceIDs, deviceID)
		messageCounts = append(messageCounts, float64(messageCount))
	}

	return deviceIDs, messageCounts
}

func createBarChart(deviceIDs []string, messageCounts []float64) {
	barChart := chart.BarChart{
		Title: "Number of Messages per device ",
		Height: 512,
		Width:  1024,
		Bars:   []chart.Value{},
	}

	for i := 0; i < len(deviceIDs); i++ {
		barChart.Bars = append(barChart.Bars, chart.Value{
			Value: messageCounts[i],
			Label: deviceIDs[i],
		})
	}

	f, err := os.Create("bar_chart.png") 
	if err != nil {
		log.Fatalf("Failed to create file: %v", err)
	}
	defer f.Close()

	err = barChart.Render(chart.PNG, f)
	if err != nil {
		log.Fatalf("Failed to render chart: %v", err)
	}

	log.Println("Bar chart saved as bar_chart.png")
}

func main() {
	db, err := sql.Open("mysql", "user:password@tcp(mysql:3306)/proxy?parseTime=true&loc=Local")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	deviceIDs, messageCounts := fetchMessageCounts(db)
	createBarChart(deviceIDs, messageCounts)
}
