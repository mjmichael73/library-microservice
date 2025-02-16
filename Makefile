down:
	@docker compose down --remove-orphans --volumes
build:
	@docker compose build
up: build
	@docker compose up -d
ps:
	@docker compose ps