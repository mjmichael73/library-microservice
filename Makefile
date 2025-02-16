down:
	@docker compose down --remove-orphans --volumes
build:
	@docker compose build
up: build
	@docker compose up -d
	sleep 5
	@docker compose exec -it userservice-db sh -c "psql -h localhost -U userservice_db_user -d userservice_db -f /data/schema.sql"
	@docker compose exec -it bookservice-db sh -c "psql -h localhost -U bookservice_db_user -d bookservice_db -f /data/schema.sql"
	@docker compose exec -it bookservice-db sh -c "psql -h localhost -U bookservice_db_user -d bookservice_db -f /data/seed.sql"
	@docker compose exec -it loanservice-db sh -c "psql -h localhost -U loanservice_db_user -d loanservice_db -f /data/schema.sql"
ps:
	@docker compose ps
runUserService:
	@docker compose exec -it userservice-app sh -c "go build -o main . && ./main"
runBookService:
	@docker compose exec -it bookservice-app sh -c "go build -o main . && ./main"
runLoanService:
	@docker compose exec -it loanservice-app sh -c "go build -o main . && ./main"