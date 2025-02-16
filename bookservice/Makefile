down:
	@docker compose down --remove-orphans --volumes
build:
	@docker compose build
up: build
	@docker compose up -d
clean: down build up
	@sleep 3
	@docker compose exec -it db sh -c "psql -h localhost -U bookservice_db_user -d bookservice_db -f /data/schema.sql"
	@docker compose exec -it db sh -c "psql -h localhost -U bookservice_db_user -d bookservice_db -f /data/seed.sql"
ps:
	@docker compose ps
run: up
	@docker compose exec -it app sh -c "go build -o main . && ./main"