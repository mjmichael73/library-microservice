up:
	@docker compose up -d
down:
	@docker compose down
clean:
	@docker compose down --remove-orphans --volumes
	@docker compose up --build -d
ps:
	@docker compose ps
dbexec:
	@docker compose exec -it db bash
dbpsql:
	@docker compose exec -it db sh -c "psql -h localhost -U userservice_db_user -d userservice_db"
dbschema:
	@docker compose exec -it db sh -c "psql -h localhost -U userservice_db_user -d userservice_db -f /data/schema.sql"
run:
	@go run main.go