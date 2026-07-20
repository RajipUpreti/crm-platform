.PHONY: up down build restart logs ps api-shell web-shell go-test go-fmt npm-install clean

up:
	docker compose -f compose.dev.yaml up -d --build

down:
	docker compose -f compose.dev.yaml down

build:
	docker compose -f compose.dev.yaml build

restart:
	docker compose -f compose.dev.yaml restart

logs:
	docker compose -f compose.dev.yaml logs -f

ps:
	docker compose -f compose.dev.yaml ps

api-shell:
	docker compose -f compose.dev.yaml exec api sh

web-shell:
	docker compose -f compose.dev.yaml exec web sh

go-test:
	docker compose -f compose.dev.yaml exec api go test ./...

go-fmt:
	docker compose -f compose.dev.yaml exec api gofmt -w .

npm-install:
	docker compose -f compose.dev.yaml exec web npm install

clean:
	docker compose -f compose.dev.yaml down -v --remove-orphans


.PHONY: migrate-up migrate-down migrate-version migrate-force migrate-create

migrate-up:
	docker compose -f compose.dev.yaml --profile tools run --rm migrate \
		-path=/migrations \
		-database="postgres://crm:crm_password@crm-db:5432/crm?sslmode=disable" \
		up

migrate-down:
	docker compose -f compose.dev.yaml --profile tools run --rm migrate \
		-path=/migrations \
		-database="postgres://crm:crm_password@crm-db:5432/crm?sslmode=disable" \
		down 1

migrate-version:
	docker compose -f compose.dev.yaml --profile tools run --rm migrate \
		-path=/migrations \
		-database="postgres://crm:crm_password@crm-db:5432/crm?sslmode=disable" \
		version

migrate-force:
	@test -n "$(VERSION)" || (echo "VERSION is required"; exit 1)
	docker compose -f compose.dev.yaml --profile tools run --rm migrate \
		-path=/migrations \
		-database="postgres://crm:crm_password@crm-db:5432/crm?sslmode=disable" \
		force $(VERSION)

migrate-create:
	@test -n "$(NAME)" || (echo "NAME is required"; exit 1)
	docker compose -f compose.dev.yaml --profile tools run --rm \
		--entrypoint migrate migrate \
		create -ext sql -dir /migrations -seq $(NAME)