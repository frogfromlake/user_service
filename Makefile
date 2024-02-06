postgres:
	@echo "Starting db_user_service..."
	docker run --name db_user_service -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:16-alpine

createdb:
	@echo "Creating database..."
	docker exec -it db_user_service createdb --username=root --owner=root streamfair_user_service_db

dropdb:
	@echo "Dropping database..."
	docker exec -it db_user_service dropdb streamfair_user_service_db

createmigration:
	@echo "Creating migration..."
	migrate create -ext sql -dir db/migration -seq init_schema

migrateup migrateup1 migratedown migratedown1:
	@echo "Migrating..."
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/streamfair_user_service_db?sslmode=disable" -verbose $(if $(filter migrateup1 migratedown1,$@),$(subst migrate,,$@),) $(if $(filter migrateup migratedown,$@),up,down) $(if $(filter migrateup1 migratedown1,$@),1,)

dbclean: migratedown migrateup
	clear

sqlc:
	sqlc generate

# testout, dbtestout, apitestout are used to redirect test output to a file
OUT ?= 0

testout: OUT=1
testout: test

dbtestout: OUT=1
dbtestout: dbtest

apitestout: OUT=1
apitestout: apitest

utiltestout: OUT=1
utiltestout: utiltest

tokentestout: OUT=1
tokentestout: tokentest

test:
	@if [ $(OUT) -eq 1 ]; then \
		go test -v -cover -count=1 ./... > tests.log; \
	else \
		go test -v -cover -count=1 ./... ; \
	fi

dbtest:
	@if [ $(OUT) -eq 1 ]; then \
		go test -v -cover -count=1 ./db/sqlc > db_tests.log; \
	else \
		go test -v -cover -count=1 ./db/sqlc ; \
	fi

apitest:
	@if [ $(OUT) -eq 1 ]; then \
		go test -v -cover -count=1 ./api > api_tests.log; \
	else \
		go test -v -cover -count=1 ./api ; \
	fi

utiltest:
	@if [ $(OUT) -eq 1 ]; then \
		go test -v -cover -count=1 ./util > util_tests.log; \
	else \
		go test -v -cover -count=1 ./util ; \
	fi

tokentest:
	@if [ $(OUT) -eq 1 ]; then \
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
	rm -f coverage.out tests.log db_tests.log api_tests.log

proto:
	rm -f pb/*.go
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
		--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
		proto/*.proto

evans:
	evans --host localhost --port 9092 -r repl

.PHONY: createdb dropdb postgres migrateup migrateup1 migratedown migratedown1 sqlc test dbtest apitest testout dbtestout apitestout utiltest utiltestout tokentest tokentestout dbclean server mock clean debug proto evans