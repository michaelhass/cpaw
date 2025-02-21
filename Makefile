APP_NAME=cpaw
BINARY_PATH=./tmp/bin/${APP_NAME}
MIGRATION_PATH=./db/migrations
CPAW_DB=cpaw.db

all: build run test migrate_create migrate_up migrate_down
.PHONY: all

build:
	templ generate
	go build -o ${BINARY_PATH}

run:
	./${BINARY_PATH}

test:
	 go test -v ./...

migrate_create:
ifdef name
	migrate create -ext sql -dir ${MIGRATION_PATH} ${name}
else
	@echo "param 'name' not defined"
endif

migrate_up:
	migrate -path ${MIGRATION_PATH} -database "sqlite3://${CPAW_DB}" -verbose up

migrate_down:
	migrate -path ${MIGRATION_PATH} -database "sqlite3://${CPAW_DB}" -verbose down

clean:
	go clean
	rm -rf bin
