.PHONY: up migrate rollback seed setup reset

up:
	docker-compose up -d

migrate:
	go run ./cmd/migrate

rollback:
	go run ./cmd/rollback

seed:
	go run ./cmd/seed

setup: up migrate seed
