package db

import (
	"database/sql"
	"log"
	"time"

	"auth-service/internal/config"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var DB *sql.DB

func ConnectToDB() {
	var err error

	log.Println("🔐 Using DB URL:", config.AppConfig.DBUrl)

	DB, err = sql.Open("pgx", config.AppConfig.DBUrl)
	if err != nil {
		log.Fatalf("❌ Failed to parse DB URL: %v", err)
	}

	DB.SetMaxOpenConns(10)
	DB.SetMaxIdleConns(5)
	DB.SetConnMaxLifetime(time.Hour)

	// Retry DB connection for up to 10 seconds
	for i := 0; i < 10; i++ {
		err = DB.Ping()
		if err == nil {
			break
		}
		log.Println("🔁 Waiting for DB to be ready...")
		time.Sleep(1 * time.Second)
	}

	if err != nil {
		log.Fatalf("❌ Failed to connect to DB after retries: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL database")
}

//Initializes the DB connection (e.g., PostgreSQL)
