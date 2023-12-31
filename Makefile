postgres:
	docker run --name postgres16 --network nex-network -p 6969:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:16-alpine

createdb:
	docker exec -it postgres16 createdb --username=root --owner=root ctt_test_001

dropdb:
	docker exec -it postgres16 dropdb ctt_test_001

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:6969/ctt_test_001?sslmode=disable" -verbose up

migrateup1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:6969/ctt_test_001?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:6969/ctt_test_001?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:6969/ctt_test_001?sslmode=disable" -verbose down 1

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

.PHONY: postgres createdb dropdb migrateup migrateup1 migratedown migratedown1 sqlc test server