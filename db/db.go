package db

import (
	"fmt"
	"os"
	"os/user"

	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq" // load Postgresql driver here, since we have no access to main
)

var db *sqlx.DB

// Connect attempts to establish a database connection
func Connect() error {

	if db != nil {
		return nil
	}

	var username = "nobody"
	var password string

	u, err := user.Current()
	if err == nil {
		username = u.Username
	}
	if os.Getenv("COCKROACH_USER") != "" {
		username = os.Getenv("COCKROACH_USER")
		password = os.Getenv("COCKROACH_PASS")
	}

	dsn := fmt.Sprintf("postgresql://%s:%s@localhost:26257/cycore?sslmode=require", username, password)
	if os.Getenv("DSN") != "" {
		dsn = os.Getenv("DSN")
	}

	db, err = sqlx.Open("postgres", dsn)
	if err != nil {
		return err
	}

	return db.Ping()
}

func ensureDatabase() {
	if err := Connect(); err != nil {
		panic("no database connection: " + err.Error())
	}
}

// Get returns a database connection handle for the database
func Get() *sqlx.DB {
	ensureDatabase()
	return db
}

// LogContact adds a log entry to the database recording a contact request
func LogContact(name, email string) error {
	ensureDatabase()

	_, err := db.Exec(`INSERT INTO contact_requests (name,email) VALUES ($1, $2)`, name, email)
	return err
}
