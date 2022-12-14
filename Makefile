DOCKER_COMPOSE_EXEC_DB=docker-compose exec db psql -U postgres -c
DOCKER_COMPOSE_RUN_GO=docker-compose run --rm golang

.PHONY: lint
lint:
	@golangci-lint run

start-local-db-server: ## Start a local postgresql instance.
	@docker-compose up -d db

stop-local-db-server: ## Stop the running postgresql instance.
	@docker-compose stop db

delete-local-db: ## Delete all data, tables and databases of the local DB.
	@echo "This will delete all existing DB data. Are you sure? [y/N]" && read ans && [ $${ans:-N} = y ]
	@$(DOCKER_COMPOSE_EXEC_DB) "drop database if exists bronze;"
	@$(DOCKER_COMPOSE_EXEC_DB) "drop database if exists silver;"

delete-local-db-no-conf:
	@$(DOCKER_COMPOSE_EXEC_DB) "drop database if exists bronze;"
	@$(DOCKER_COMPOSE_EXEC_DB) "drop database if exists silver;"

create-local-db: ## Create the necessary databases in the local DB.
	@$(DOCKER_COMPOSE_EXEC_DB) "create database bronze;"
	@$(DOCKER_COMPOSE_EXEC_DB) "create database silver;"

open-local-db-terminal: ## Open a psql connected to the local DB.
	@docker-compose exec db psql -U postgres

migrate-local-db: ## Run schema migrations in the local DB.
	@$(DOCKER_COMPOSE_RUN_GO) go run cmd/migrate/migrate.go

location: ## Run script to fetch location data from IBGE and save to the local DB.
	@$(DOCKER_COMPOSE_RUN_GO) go run cmd/locations/locations.go

indicators: ## Run script to fetch indicators data from IBGE and save to the local DB.
	@$(DOCKER_COMPOSE_RUN_GO) go run cmd/indicators/indicators.go

polling-places-info: ## Run script to fetch polling places and ballot boxes data from TSE and save to the local DB.
	@$(DOCKER_COMPOSE_RUN_GO) go run cmd/polling-places-info/polling.go

ballot-box-reports: ## Run script to fetch ballot box reports data from TSE and save to the local DB.
	@$(DOCKER_COMPOSE_RUN_GO) go run cmd/ballotbox-reports/reports.go

section-results: ## Run script to fetch section results from TSE and save to the local DB.
	@$(DOCKER_COMPOSE_RUN_GO) go run cmd/section-results/results.go

from-scratch: start-local-db-server delete-local-db-no-conf create-local-db migrate-local-db location indicators polling-places-info ballot-box-reports section-results ## Delete the current databases, recreate and populate them with data.

bronze: location indicators polling-places-info ballot-box-reports section-results ## Populate bronze database with raw data from different sources.

.PHONY: build
.SILENT: help
help: ## Show this help message
	set -x
	echo "Usage: make [target] ..."
	echo ""
	echo "Available targets:"
	grep ':.* ##\ ' ${MAKEFILE_LIST} | awk '{gsub(":[^#]*##","\t"); print}' | column -t -c 2 -s $$'\t' | sort
