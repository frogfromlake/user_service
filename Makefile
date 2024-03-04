###  VARIABLES  ###
ifndef GITHUB_ACTIONS
-include ./env/db.env
endif

# Docker Network
DB_NETWORK := db_access_network
SERVICE_NETWORK := streamfair_internal_network

# Service Container
## ADJUST FOR EACH SERVICE ##
SERVICE_IMAGE := streamfair_user_svc
SERVICE_TAG := latest

GRPC_PORT := 9094
GRPC_HOST_PORT := 9590

GRPC_GATEWAY_PORT := 8084
GRPC_GATEWAY_HOST_PORT := 8580

# Database Container
## ADJUST FOR EACH SERVICE ##
DB_IMAGE := postgres:16-alpine
DB_CONTAINER_NAME := db_user_service

DB_PORT := 5432
DB_HOST_PORT := 5436

DB_NAME := $(or $(DB_NAME),$(shell grep POSTGRES_DB ./env/db.env | cut -d '=' -f2))
DB_USER := $(or $(DB_USER),$(shell grep POSTGRES_USER ./env/db.env | cut -d '=' -f2))
DB_PASSWORD := $(or $(DB_PASSWORD),$(shell grep POSTGRES_PASSWORD ./env/db.env | cut -d '=' -f2))

DB_HOST := localhost
DB_SOURCE_SERVICE := "postgresql://${DB_USER}:${DB_PASSWORD}@${DOCKER_NETWORK}:${DB_HOST_PORT}/${DB_NAME}?sslmode=disable"
DB_SOURCE_MIGRATION := "postgresql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_HOST_PORT}/${DB_NAME}?sslmode=disable"

# Migrate
MIGRATION_DIR := db/migration
MIGRATION_NAME := init_schema

# Proto
## ADJUST FOR EACH SERVICE ##
PROTO_DIR := proto
PB_DIR := pb
USER_DIR := user

# Test
TEST_DIR := ./...
DB_TEST_DIR := ./db/sqlc
API_TEST_DIR := ./api
UTIL_TEST_DIR := ./util
SERVER_TEST_DIR := ./gapi
# Test Output
TEST_FILE := tests.log
DB_TEST_FILE := db_tests.log
API_TEST_FILE := api_tests.log
UTIL_TEST_FILE := util_tests.log
SERVER_TEST_FILE := server_tests.log
# Output Flag
OUT ?= 0

# Go
ENTRY_POINT := main.go

# Mock-gen
MOCK_SOURCE := db/sqlc/store.go
MOCK_DEST := db/mock/store_mock.go

# Documentation
SWAGGER_DIR := doc/swagger
SWAGGER_DOC_NAME := streamfair_user_service


###  TARGETS  ###
# Network Management
network:
	@docker network inspect ${DB_NETWORK} >/dev/null 2>&1 || \
	docker network create --driver bridge ${DB_NETWORK}
	@docker network inspect ${SERVICE_NETWORK} >/dev/null 2>&1 || \
	docker network create --driver bridge ${SERVICE_NETWORK}

# Service Management
service_image:
	@docker build -t ${SERVICE_IMAGE}:${SERVICE_TAG} .

service_container:
	@docker run --name ${SERVICE_IMAGE} --network ${DB_NETWORK} --network ${SERVICE_NETWORK} -p ${GRPC_GATEWAY_HOST_PORT}:${GRPC_GATEWAY_PORT} -p ${GRPC_HOST_PORT}:${GRPC_PORT} -e DB_SOURCE=${DB_SOURCE_SERVICE} ${SERVICE_IMAGE}:${SERVICE_TAG}

# DB Management
db_container: network
	@docker start ${DB_CONTAINER_NAME} >/dev/null 2>&1 || \
	docker run --name ${DB_CONTAINER_NAME} --network ${DB_NETWORK} -p ${DB_HOST_PORT}:${DB_PORT} -e POSTGRES_USER=${DB_USER} -e POSTGRES_PASSWORD=${DB_PASSWORD} -d ${DB_IMAGE}

createdb:
	@sleep 1
	@docker exec -it ${DB_CONTAINER_NAME} psql -U ${DB_USER} -lqt | cut -d \| -f 1 | grep -qw ${DB_NAME} >/dev/null 2>&1 || \
	docker exec -it ${DB_CONTAINER_NAME} createdb --username=${DB_USER} --owner=${DB_USER} ${DB_NAME}

dropdb:
	@docker exec -it ${DB_CONTAINER_NAME} dropdb ${DB_NAME}

createmigration:
	migrate create -ext sql -dir ${MIGRATION_DIR} -seq ${MIGRATION_NAME}

migrateup:
	@migrate -path ${MIGRATION_DIR} -database ${DB_SOURCE_MIGRATION} -verbose up

migrateup1:
	@migrate -path ${MIGRATION_DIR} -database ${DB_SOURCE_MIGRATION} -verbose up 1

migratedown:
	@migrate -path ${MIGRATION_DIR} -database ${DB_SOURCE_MIGRATION} -verbose down

migratedown1:
	@migrate -path ${MIGRATION_DIR} -database ${DB_SOURCE_MIGRATION} -verbose down 1

dbclean: migratedown migrateup
	clear


# Execution
server: network db_container createdb migrateup
	@go run ${ENTRY_POINT}

# Cleanup
down:
	@if [ "$$(docker ps -aq -f name=$(DB_CONTAINER_NAME))" ]; then \
		docker stop $(DB_CONTAINER_NAME); \
		docker rm $(DB_CONTAINER_NAME); \
	fi
	@if [ "$$(docker ps -aq -f name=$(SERVICE_IMAGE))" ]; then \
		docker stop $(SERVICE_IMAGE); \
		docker rm $(SERVICE_IMAGE); \
	fi
	@if [ "$$(docker network ls -q -f name=$(DB_NETWORK))" ]; then \
		docker network rm $(DB_NETWORK); \
	fi
	@if [ "$$(docker network ls -q -f name=$(SERVICE_NETWORK))" ]; then \
		docker network rm $(SERVICE_NETWORK); \
	fi


# SQLC Generation
sqlc:
	sqlc generate


# Mock Generation
mock:
	mockgen -source=${MOCK_SOURCE} -destination=${MOCK_DEST}


# Proto Generation
## ADJUST FOR EACH SERVICE ##
proto: proto_core

proto_core: clean_pb proto_user
	protoc --proto_path=${PROTO_DIR} --go_out=${PB_DIR} --go_opt=paths=source_relative \
		--go-grpc_out=${PB_DIR} --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=${PB_DIR} --grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=$(SWAGGER_DIR) --openapiv2_opt=allow_merge=true,merge_file_name=${SWAGGER_DOC_NAME},preserve_rpc_order=true \
		${PROTO_DIR}/*.proto
		statik -src=./$(SWAGGER_DIR) -dest=./doc

proto_user: clean_user_dir
	protoc --proto_path=${PROTO_DIR} --go_out=${PB_DIR} --go_opt=paths=source_relative \
		--go-grpc_out=${PB_DIR} --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=${PB_DIR} --grpc-gateway_opt=paths=source_relative \
		${PROTO_DIR}/$(USER_DIR)/*.proto

clean_pb:
	rm -f $(PB_DIR)/*.go
	rm -f $(SWAGGER_DIR)/*.swagger.json
	
clean_user_dir:
	rm -f $(USER_DIR)/*.go


# Evans GRPC Client
evans:
	evans --host ${DB_HOST} --port ${GRPC_PORT} -r repl


# Testing
test:
	@if [ $(OUT) -eq 1 ]; then \
		go test -v -cover ${TEST_DIR} > ${TEST_FILE}; \
	else \
		go test -v -cover ${TEST_DIR} ; \
	fi

dbtest:
	@if [ $(OUT) -eq 1 ]; then \
		go test -v -cover ${DB_TEST_DIR} > ${DB_TEST_FILE}; \
	else \
		go test -v -cover ${DB_TEST_DIR} ; \
	fi

apitest:
	@if [ $(OUT) -eq 1 ]; then \
		go test -v -cover ${API_TEST_DIR} > ${API_TEST_FILE}; \
	else \
		go test -tags=-coverage -v -cover ${API_TEST_DIR} ; \
	fi

utiltest:
	@if [ $(OUT) -eq  1 ]; then \
		go test -v -cover -count=1 ${UTIL_TEST_DIR} > ${UTIL_TEST_FILE}; \
	else \
		go test -v -cover -count=1 ${UTIL_TEST_DIR} ; \
	fi

servertest:
	@if [ $(OUT) -eq  1 ]; then \
		go test -v -cover -count=1 ${SERVER_TEST_DIR} > ${SERVER_TEST_FILE}; \
	else \
		go test -v -cover -count=1 ${SERVER_TEST_DIR} ; \
	fi

coverage_html:
	go test -coverprofile=coverage.out ${TEST_DIR}
	go tool cover -html=coverage.out
	rm coverage.out

clean:
	rm -f *_tests.log


# PHONY Targets
.PHONY: network db_container createdb dropdb createmigration migrateup migrateup1 migratedown migratedown1 dbclean service_image service_container server sqlc mock proto proto_core proto_user clean_pb clean_user_dir evans test dbtest apitest utiltest servertest coverage_html clean
