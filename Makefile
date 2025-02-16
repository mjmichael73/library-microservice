down:
	@docker compose down --remove-orphans --volumes
build:
	@docker compose build
up: build
	@docker compose up -d
ps:
	@docker compose ps
runUserService:
	@docker compose exec -it userservice-app sh -c "go build -o main . && ./main"
runBookService:
	@docker compose exec -it bookservice-app sh -c "go build -o main . && ./main"
runLoanService:
	@docker compose exec -it loanservice-app sh -c "go build -o main . && ./main"