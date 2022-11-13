DOCKER_COMPOSE_EXEC_DB=docker-compose exec db psql -U postgres -c
DOCKER_COMPOSE_RUN_GO=docker-compose run --rm golang

start-local-db-server: ## Start a local postgresql instance.
	@docker-compose up db -d

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

population: ## Run script to fetch population data from IBGE and save to the local DB.
	@$(DOCKER_COMPOSE_RUN_GO) go run cmd/populations/populations.go

from-scratch: start-local-db-server delete-local-db-no-conf create-local-db migrate-local-db location population

.PHONY: build
.SILENT: help
help: ## Show this help message
	set -x
	echo "Usage: make [target] ..."
	echo ""
	echo "Available targets:"
	grep ':.* ##\ ' ${MAKEFILE_LIST} | awk '{gsub(":[^#]*##","\t"); print}' | column -t -c 2 -s $$'\t' | sort
