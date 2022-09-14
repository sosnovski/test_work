APP = mascot


test: ## run unit tests
	go test ./internal/...

env: ## generate sample env file
	touch .env
	@echo "\
MASCOT_ADDR=:8080\n\
MASCOT_SEAMLESS_URI=/mascot/seamless\n\
MASCOT_POSTGRES_DSN=postgresql://localhost/mascot?user=mascot&password=mascot&sslmode=disable\n" > .env

envup: ## local environment up
	docker-compose -p $(APP)-env -f ./local/docker-compose.base.yml up --remove-orphans

envdown:: ## local environment up
	docker-compose -p $(APP)-env -f ./local/docker-compose.base.yml down