start-local-db-server:
	@docker-compose up db -d

stop-local-db-server:
	@docker-compose stop db

delete-local-db:
	@echo "This will delete all existing DB data. Are you sure? [y/N]" && read ans && [ $${ans:-N} = y ]
	@docker-compose exec db psql -U postgres -c "drop database if exists bronze;"
	@docker-compose exec db psql -U postgres -c "drop database if exists silver;"

create-local-db:
	@docker-compose exec db psql -U postgres -c "create database bronze;"
	@docker-compose exec db psql -U postgres -c "create database silver;"

open-local-db-terminal:
	@docker-compose exec db psql -U postgres

migrate-local-db:
	@docker-compose run --rm golang go run cmd/migrate/migrate.go

populate-locations:
	@docker-compose run --rm golang go run cmd/locations/locations.go