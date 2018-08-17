
build:
	docker-compose build

server:
	UID=${UID} GID=${GID} docker-compose up server

client:
	UID=${UID} GID=${GID} docker-compose up client

clean:
	-rm sessions/*
	docker-compose down
