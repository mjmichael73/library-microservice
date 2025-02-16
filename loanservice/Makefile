down:
	@docker compose down --remove-orphans --volumes
build:
	@docker compose build
up: build
	@docker compose up -d
clean: down build up
	@sleep 3
	@docker compose exec -it db sh -c "psql -h localhost -U loanservice_db_user -d loanservice_db -f /data/schema.sql"
ps:
	@docker compose ps
run: up
	@docker compose exec -it app sh -c "go build -o main . && ./main"