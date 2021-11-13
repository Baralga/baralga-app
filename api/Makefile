.PHONY: clean test security build run swag

APP_NAME = baralga_backend
BUILD_DIR = $(PWD)/build
MIGRATIONS_FOLDER = $(PWD)/migrations
DATABASE_URL = postgres://postgres:postgres@localhost:5432/baralga?sslmode=disable

clean:
	rm -rf ./build

linter:
	golangci-lint run

test:
	go test -v -timeout 30s -coverprofile=cover.out -cover ./...
	go tool cover -func=cover.out

build: clean test
	CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BUILD_DIR)/$(APP_NAME) .

migrate.up:
	migrate -path $(MIGRATIONS_FOLDER) -database "$(DATABASE_URL)" up

migrate.down:
	migrate -path $(MIGRATIONS_FOLDER) -database "$(DATABASE_URL)" down

migrate.drop:
	migrate -path $(MIGRATIONS_FOLDER) -database "$(DATABASE_URL)" drop

migrate.force:
	migrate -path $(MIGRATIONS_FOLDER) -database "$(DATABASE_URL)" force $(version)

docker.postgres:
	docker-compose up
