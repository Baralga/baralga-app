.PHONY: clean test security build run swag

APP_NAME = baralga
BUILD_DIR = $(PWD)/build
MIGRATIONS_FOLDER = $(PWD)/shared/migrations
DATABASE_URL = postgres://postgres:postgres@localhost:5432/baralga?sslmode=disable

clean:
	rm -rf ./build

linter:
	golangci-lint run

arch-go.install:
	go install github.com/fdaines/arch-go@v0.8.6

arch-go.check:
	arch-go --verbose

test:
	go test -v -timeout 60s -coverprofile=cover.out -cover ./...
	go tool cover -func=cover.out

build: clean
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

app.yaml: .ci-util/app.tpl.yaml .ci-util/generate-gcloud-app.go
	go run .ci-util/generate-gcloud-app.go

release.test:
	goreleaser release --snapshot --rm-dist