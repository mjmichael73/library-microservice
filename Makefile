down:
	@docker compose down --remove-orphans --volumes
build:
	@docker compose build
up: build
	@docker compose up -d
db:
	@docker compose exec -it userservice-db sh -c "psql -h localhost -U userservice_db_user -d userservice_db -f /data/schema.sql"
	@docker compose exec -it userservice-db sh -c "psql -h localhost -U userservice_db_user -d userservice_db -f /data/seed.sql"
	@docker compose exec -it bookservice-db sh -c "psql -h localhost -U bookservice_db_user -d bookservice_db -f /data/schema.sql"
	@docker compose exec -it bookservice-db sh -c "psql -h localhost -U bookservice_db_user -d bookservice_db -f /data/seed.sql"
	@docker compose exec -it loanservice-db sh -c "psql -h localhost -U loanservice_db_user -d loanservice_db -f /data/schema.sql"
ps:
	@docker compose ps