package main

import (
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	m, err := migrate.New(
		"file://./migrations",
		"postgres://postgres:postgres@localhost:5432/eleicoes?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	err = m.Up() // or m.Step(2) if you want to explicitly set the number of migrations to run
	if err != nil {
		log.Fatal(err)
	}

}
