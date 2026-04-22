.PHONY: test test-python test-go test-ts lint up down build clean

test: test-python test-go test-ts

test-python:
	cd services/analytics-engine && pip install -r requirements.txt -q && pytest -v

test-go:
	cd services/ingestion-gateway && go test -v ./...

test-ts:
	cd services/dashboard-api && npm install --silent && npm test

lint: lint-python lint-go lint-ts

lint-python:
	cd services/analytics-engine && flake8 --max-line-length=120 app.py

lint-go:
	cd services/ingestion-gateway && go vet ./...

lint-ts:
	cd services/dashboard-api && npm run lint

up:
	docker compose up -d --build

down:
	docker compose down

build:
	docker compose build

clean:
	docker compose down -v --rmi local
