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