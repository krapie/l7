build:
	CGO_ENABLED=0 go build -o ./bin/plumber .

docker-build:
	docker build --push -t krapi0314/plumber .

docker-compose-up:
	docker-compose -f ./docker/docker-compose.yml up --build -d

docker-compose-down:
	docker-compose -f ./docker/docker-compose.yml down
