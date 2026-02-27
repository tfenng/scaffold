DB_URL=postgres://user:pass@localhost:5432/app?sslmode=disable

sqlc:
	sqlc generate

migrate-up:
	migrate -path migrations -database "$(DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(DB_URL)" down 1

migrate-new:
	migrate create -ext sql -dir migrations -seq $(name)

.PHONY: sqlc migrate-up migrate-down migrate-new
