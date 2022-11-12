package main

import (
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

type logger struct {
	*log.Logger
}

func (l *logger) Verbose() bool {
	return false
}

func main() {
	migrateDB("bronze")
	// TODO: Create silver and gold databases
}

func migrateDB(name string) {
	log.Println("Migrating database", name)

	m, err := migrate.New(
		fmt.Sprintf("file://./migrations/%s", name),
		fmt.Sprintf("postgres://postgres:postgres@db.local:5432/%s?sslmode=disable", name))
	if err != nil {
		log.Fatal(err)
	}

	m.Log = &logger{
		Logger: log.Default(),
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatal(err)
	}

	if err != nil && errors.Is(err, migrate.ErrNoChange) {
		log.Println("No change.")
		return
	}

	log.Println("Done.")
}
