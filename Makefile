migrate-local-up: 
	GOOSE_DRIVER=postgres GOOSE_DBSTRING="host=localhost port=5432 user=bot_user password=bot_password dbname=bot_dev sslmode=disable" goose -dir=${PWD}/migrations up

migrate-local-down: 
	GOOSE_DRIVER=postgres GOOSE_DBSTRING="host=localhost port=5432 user=bot_user password=bot_password dbname=bot_dev sslmode=disable" goose -dir=${PWD}/migrations down
