include .env
export

migrate-up:
	migrate -database "$(DATABASE_URL)" -path migrations up

migrate-down:
	migrate -database "$(DATABASE_URL)" -path migrations down

migrate-version:
	migrate -database "$(DATABASE_URL)" -path migrations version

migrate-force-0:
	migrate -database "$(DATABASE_URL)" -path migrations force 0