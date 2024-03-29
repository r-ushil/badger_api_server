export MONGO_SCHEME		:= mongodb+srv
export MONGO_HOST			:= badgerdbdev.syhi5.mongodb.net
export MONGO_PATH			:= test
export MONGO_USER			:= coleoidea
export MONGO_PASS			:= MzCuyaDKrPV7Jtb6
export MONGO_DB_NAME	:= badger_db
export PORT						:= 3000

export GOOGLE_APPLICATION_CREDENTIALS := ./google-service-account.json

run: build
		/usr/local/bin/badger_api

build:
		go build -o /usr/local/bin/badger_api ./cmd

generate:
		buf generate

clean_generate:
		rm -rf ./buf && buf generate

integration_tests:
		go test -v ./test/integration/...

unit_tests:
		go test -v ./test/unit/...