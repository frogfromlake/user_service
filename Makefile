# Variables
GRPC_PORT :=   9094
DB_PORT :=   5436
DB_CONTAINER_NAME := db_user_service
DB_NAME := streamfair_user_service_db
DB_USER := root
DB_PASSWORD := secret
DB_HOST := localhost

PROTO_DIR := proto
PB_DIR := pb
USER_DIR := user

OUT ?=   0

# Targets
postgres:
	@echo "Starting ${DB_CONTAINER_NAME}..."
	docker run --name ${DB_CONTAINER_NAME} -p ${DB_PORT}:5432 -e POSTGRES_USER=${DB_USER} -e POSTGRES_PASSWORD=${DB_PASSWORD} -d postgres:16-alpine

createdb:
	@echo "Creating database..."
	docker exec -it ${DB_CONTAINER_NAME} createdb --username=${DB_USER} --owner=${DB_USER} ${DB_NAME}

dropdb:
	@echo "Dropping database..."
	docker exec -it ${DB_CONTAINER_NAME} dropdb ${DB_NAME}

createmigration:
	@echo "Creating migration..."
	migrate create -ext sql -dir db/migration -seq init_schema

migrateup migrateup1 migratedown migratedown1:
	@echo "Migrating..."
	migrate -path db/migration -database "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable" -verbose $(if $(filter migrateup1 migratedown1,$@),$(subst migrate,,$@),) $(if $(filter migrateup migratedown,$@),up,down) $(if $(filter migrateup1 migratedown1,$@),1,)

dbclean: migratedown migrateup
	clear

sqlc:
	sqlc generate

test:
	@if [ $(OUT) -eq   1 ]; then \
		go test -v -cover -count=1 ./... > tests.log; \
	else \
		go test -v -cover -count=1 ./... ; \
	fi

dbtest:
	@if [ $(OUT) -eq   1 ]; then \
		go test -v -cover -count=1 ./db/sqlc > db_tests.log; \
	else \
		go test -v -cover -count=1 ./db/sqlc ; \
	fi

apitest:
	@if [ $(OUT) -eq   1 ]; then \
		go test -v -cover -count=1 ./api > api_tests.log; \
	else \
		go test -v -cover -count=1 ./api ; \
	fi

utiltest:
	@if [ $(OUT) -eq   1 ]; then \
		go test -v -cover -count=1 ./util > util_tests.log; \
	else \
		go test -v -cover -count=1 ./util ; \
	fi

tokentest:
	@if [ $(OUT) -eq   1 ]; then \
		go test -v -cover -count=1 ./token > token_tests.log; \
	else \
		go test -v -cover -count=1 ./token ; \
	fi

coverage_html:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out
	rm coverage.out

server:
	@go run main.go

mock:
	mockgen -source=db/sqlc/store.go -destination=db/mock/store_mock.go

clean:
	rm -f coverage.out tests.log db_tests.log api_tests.log util_tests.log token_tests.log

proto: proto_core

proto_core: clean_pb proto_user
	protoc --proto_path=${PROTO_DIR} --go_out=${PB_DIR} --go_opt=paths=source_relative \
		--go-grpc_out=${PB_DIR} --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=${PB_DIR} --grpc-gateway_opt=paths=source_relative \
		${PROTO_DIR}/*.proto

proto_user: clean_user_dir
	protoc --proto_path=${PROTO_DIR} --go_out=${PB_DIR} --go_opt=paths=source_relative \
		--go-grpc_out=${PB_DIR} --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=${PB_DIR} --grpc-gateway_opt=paths=source_relative \
		${PROTO_DIR}/$(USER_DIR)/*.proto

clean_pb:
	rm -f $(PB_DIR)/*.go

clean_user_dir:
	rm -f $(USER_DIR)/*.go

evans:
	evans --host ${DB_HOST} --port ${GRPC_PORT} -r repl

# PHONY Targets
.PHONY: postgres createdb dropdb createmigration migrateup migrateup1 migratedown migratedown1 sqlc test dbtest apitest utiltest tokentest coverage_html server mock clean proto evans proto_core proto_user clean_pb clean_user_dir
