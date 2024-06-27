package main

import (
	"database/sql"
	"log"
	"testing"
	"time"

	"github.com/Masterminds/squirrel"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// Define the User struct
type User struct {
	ID    int
	Name  string
	Email string
}

var db *sql.DB
var gormDB *gorm.DB

func init() {
	var err error

	// Initialize sql.DB
	db, err = sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/benchmark")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize gorm.DB
	gormDB, err = gorm.Open(mysql.Open("root:password@tcp(127.0.0.1:3306)/benchmark"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
}

// Benchmark for Raw SQL
func BenchmarkRawSQL(b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var user User
		query := "SELECT id, name, email FROM users WHERE id = ?"
		err := db.QueryRow(query, 1).Scan(&user.ID, &user.Name, &user.Email)
		if err != nil {
			b.Error(err)
		}
	}
	b.StopTimer()
}

func BenchmarkSquirrel(b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		query, args, err := squirrel.Select("id", "name", "email").
			From("users").
			Where(squirrel.Eq{"id": 1}).
			ToSql()

		if err != nil {
			b.Error(err)
		}

		var user User
		err = db.QueryRow(query, args...).Scan(&user.ID, &user.Name, &user.Email)
		if err != nil {
			b.Error(err)
		}
	}
	b.StopTimer()
}

// Benchmark for GORM
func BenchmarkGORM(b *testing.B) {
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		var user User
		err := gormDB.First(&user, 1).Error
		if err != nil {
			b.Error(err)
		}
	}
	b.StopTimer()
}

// Measure latency for each method
func measureLatency(b *testing.B, queryFunc func() error) time.Duration {
	start := time.Now()
	for n := 0; n < b.N; n++ {
		err := queryFunc()
		if err != nil {
			b.Error(err)
		}
	}
	elapsed := time.Since(start)
	return elapsed / time.Duration(b.N)
}

// Benchmark Latency for Raw SQL
func BenchmarkLatencyRawSQL(b *testing.B) {
	latency := measureLatency(b, func() error {
		var user User
		query := "SELECT id, name, email FROM users WHERE id = ?"
		return db.QueryRow(query, 1).Scan(&user.ID, &user.Name, &user.Email)
	})
	b.Log("Raw SQL Latency:", latency)
}

// Benchmark Latency for Squirrel
func BenchmarkLatencySquirrel(b *testing.B) {
	latency := measureLatency(b, func() error {
		query, args, err := squirrel.Select("id", "name", "email").
			From("users").
			Where(squirrel.Eq{"id": 1}).
			ToSql()
		if err != nil {
			return err
		}
		var user User
		return db.QueryRow(query, args...).Scan(&user.ID, &user.Name, &user.Email)
	})
	b.Log("Squirrel Latency:", latency)
}

// Benchmark Latency for GORM
func BenchmarkLatencyGORM(b *testing.B) {
	latency := measureLatency(b, func() error {
		var user User
		return gormDB.First(&user, 1).Error
	})
	b.Log("GORM Latency:", latency)
}
