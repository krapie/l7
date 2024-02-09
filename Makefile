.PHONY build:
	CGO_ENABLED=0 go build -o ./bin/plumber .

.PHONY docker-build:
	docker build --push -t krapi0314/plumber .

.PHONY docker-compose-up:
	docker-compose -f ./docker/docker-compose.yml up --build -d

.PHONY docker-compose-down:
	docker-compose -f ./docker/docker-compose.yml down

.PHONY docker-compose-yorkie-up:
	docker-compose -f ./docker/docker-compose-yorkie.yml up --build -d

.PHONY docker-compose-yorkie-down:
	docker-compose -f ./docker/docker-compose-yorkie.yml down