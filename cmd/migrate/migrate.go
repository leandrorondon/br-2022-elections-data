package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

type logger struct {
	*log.Logger
}

func (l *logger) Verbose() bool {
	return false
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal(err)
	}

	migrateDB("bronze")
	// TODO: Create silver and gold databases
}

func migrateDB(name string) {
	log.Println("Migrating database", name)

	dbURL := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOSTNAME"),
		os.Getenv("DB_PORT"),
		name,
	)

	m, err := migrate.New(
		fmt.Sprintf("file://./migrations/%s", name),
		dbURL,
	)
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
