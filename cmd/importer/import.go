package main

import (
	"bufio"
	"fmt"
	"github.com/go-kit/log"
	"github.com/go-kit/log/level"
	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
	"os"
	"strings"
)

func main() {
	l := log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))
	l = level.NewFilter(l, level.AllowInfo())
	l = log.With(l, "ts", log.DefaultTimestampUTC, "caller", log.DefaultCaller)

	db, err := sqlx.Open("sqlite3", "../../webhook-receiver.db")
	if err != nil {
		level.Error(l).Log("msg", "error opening database", "err", err)
		return
	}
	if err := db.Ping(); err != nil {
		level.Error(l).Log("msg", "error pinging database", "err", err)
		return
	}
	fmt.Println("open")
	f, err := os.OpenFile("../../cache_migrate", os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		level.Error(l).Log("err", err)
		return
	}
	defer f.Close()

	// Import all legacy cache keys into the new system
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if scanner.Text() != "" {
			parts := strings.Split(scanner.Text(), ":")
			_, err := db.NamedExec(`INSERT INTO cache (key, notification_service, date) VALUES (:key,:notification_service,:date)`,
				map[string]interface{}{
					"key":                  parts[1],
					"notification_service": "twitter",
					"date":                 parts[0],
				})
			if err != nil {
				level.Error(l).Log("msg", "error inserting row into db", "row", parts[0], "err", err)
				continue
			}
		}
	}
}
