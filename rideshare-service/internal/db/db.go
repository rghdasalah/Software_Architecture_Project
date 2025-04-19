package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/Ahmed-Abbas-2077/rideshare-service/internal/config"
	_ "github.com/lib/pq"
)

// DBManager manages connections to primary and replica databases
type DBManager struct {
	primary      *sql.DB
	replicas     []*sql.DB
	replicaIndex int
	mu           sync.Mutex
}

// NewDBManager creates a new database manager with connection pooling
func NewDBManager(cfg *config.DatabaseConfig) (*DBManager, error) {
	// Connect to primary database
	primaryDB, err := connectToDB(cfg.Primary)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to primary database: %v", err)
	}

	// Configure connection pool
	primaryDB.SetMaxOpenConns(cfg.MaxOpenConns)
	primaryDB.SetMaxIdleConns(cfg.MaxIdleConns)
	primaryDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

	// Connect to replica databases
	var replicas []*sql.DB
	for _, replicaCfg := range cfg.Replicas {
		replicaDB, err := connectToDB(replicaCfg)
		if (err != nil) {
			log.Printf("Warning: failed to connect to replica db %s:%s: %v", 
				replicaCfg.Host, replicaCfg.Port, err)
			continue
		}

		// Configure connection pool for replica
		replicaDB.SetMaxOpenConns(cfg.MaxOpenConns)
		replicaDB.SetMaxIdleConns(cfg.MaxIdleConns)
		replicaDB.SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second)

		replicas = append(replicas, replicaDB)
	}

	// Handle case when no replicas are available
	if len(replicas) == 0 && len(cfg.Replicas) > 0 {
		return nil, errors.New("failed to connect to any replica databases")
	}

	return &DBManager{
		primary:  primaryDB,
		replicas: replicas,
	}, nil
}

// NewDBManagerFromDB creates a new database manager from an existing database connection
func NewDBManagerFromDB(db *sql.DB) (*DBManager, error) {
	if db == nil {
		return nil, errors.New("nil database connection provided")
	}
	
	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}
	
	return &DBManager{
		primary:  db,
		replicas: []*sql.DB{}, // No replicas when created from a single connection
	}, nil
}

// GetPrimary returns the primary database connection for write operations
func (m *DBManager) GetPrimary() *sql.DB {
	return m.primary
}

// GetReplica returns a read-only database connection using round-robin selection
func (m *DBManager) GetReplica() *sql.DB {
	// If no replicas are available, return primary
	if len(m.replicas) == 0 {
		return m.primary
	}

	// Round-robin selection of replica
	m.mu.Lock()
	defer m.mu.Unlock()
	
	db := m.replicas[m.replicaIndex]
	m.replicaIndex = (m.replicaIndex + 1) % len(m.replicas)
	return db
}

// Close closes all database connections
func (m *DBManager) Close() {
	if m.primary != nil {
		m.primary.Close()
	}

	for _, replica := range m.replicas {
		replica.Close()
	}
}

// ExecContext executes a query on the primary database
func (m *DBManager) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return m.primary.ExecContext(ctx, query, args...)
}

// QueryContext executes a query on a replica database
func (m *DBManager) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return m.GetReplica().QueryContext(ctx, query, args...)
}

// QueryRowContext executes a query on a replica database that returns a single row
func (m *DBManager) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return m.GetReplica().QueryRowContext(ctx, query, args...)
}

// connectToDB establishes a connection to a database
func connectToDB(dbConfig config.DBConnection) (*sql.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.User, dbConfig.Password, dbConfig.DBName, dbConfig.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}