package schema

import (
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	// Blank import justified for migrating test database
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

const dest = "/migrations"

// Migrate runs the latest migration
func Migrate(url string) error {
	src := fmt.Sprintf("file://%s%s", PWD(), dest)
	m, err := migrate.New(src, url)
	if err != nil {
		log.Fatal(err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}
	return nil
}
