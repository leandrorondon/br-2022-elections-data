DOCKER_COMPOSE_EXEC_DB=docker-compose exec db psql -U postgres -c
DOCKER_COMPOSE_RUN_GO=docker-compose run --rm golang

start-local-db-server:
	@docker-compose up db -d

stop-local-db-server:
	@docker-compose stop db

delete-local-db:
	@echo "This will delete all existing DB data. Are you sure? [y/N]" && read ans && [ $${ans:-N} = y ]
	@$(DOCKER_COMPOSE_EXEC_DB) "drop database if exists bronze;"
	@$(DOCKER_COMPOSE_EXEC_DB) "drop database if exists silver;"

delete-local-db-no-conf:
	@$(DOCKER_COMPOSE_EXEC_DB) "drop database if exists bronze;"
	@$(DOCKER_COMPOSE_EXEC_DB) "drop database if exists silver;"

create-local-db:
	@$(DOCKER_COMPOSE_EXEC_DB) "create database bronze;"
	@$(DOCKER_COMPOSE_EXEC_DB) "create database silver;"

open-local-db-terminal:
	@docker-compose exec db psql -U postgres

migrate-local-db:
	@$(DOCKER_COMPOSE_RUN_GO) go run cmd/migrate/migrate.go

populate-locations:
	@$(DOCKER_COMPOSE_RUN_GO) go run cmd/locations/locations.go

populate-populations:
	@$(DOCKER_COMPOSE_RUN_GO) go run cmd/populations/populations.go

from-scratch: start-local-db-server delete-local-db-no-conf create-local-db migrate-local-db populate-locations populate-populations