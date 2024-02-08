# plumber

Plumber is a L7 load balancer from scratch in Go

## Installation

```bash
# Clone the repository
git clone https://github.com/krapie/plumber.git

# Build the binary
cd plumber
make build
```

## Usage

Basic usage:

```bash
# Start the load balancer
./bin/plumber
```

Basic usage with docker-compose:

```bash
# Test the load balancer with docker-compose
make docker-compose-up

# Start the load balancer with target backend image for service discovery
./bin/plumber --target-backend-image traefik/whoami

# Send a request to the load balancer
chmod +x ./scripts/lb_distribution_test.sh
./scripts/lb_distribution_test.sh

# Cleanup the docker-compose
make docker-compose-down
```

Yorkie usage with docker-compose:

```bash
# Test the load balancer with docker-compose
make docker-compose-yorkie-up

# Start the load balancer with target backend image for service discovery
./bin/plumber --target-backend-image yorkieteam/yorkie

# Test with yorkie-js-sdk
git clone https://github.com/yorkie-team/yorkie-js-sdk.git
cd yorkie-js-sdk
npm install
sed -i html 's#http://localhost:8080#http://localhost#g' ./public/index.html
npm run dev

# Cleanup the docker-compose
make docker-compose-yorkie-down
```

## Roadmap

Plumber aims to support [Yorkie](https://github.com/yorkie-team/yorkie) as a backend for the load balancer.
The following features are planned to be implemented first:

### v0.1.0

- [x] Support static load balancing with round-robin algorithm
- [x] Support backends health check 

### v0.2.0

- [x] Support consistent hashing algorithm with maglev
- [x] Support backend service discovery with Docker API
- [x] Support mechanism to resolve split-brain of long-lived connection

### v0.3.0

- [ ] Support interceptor to modify request/response
- [ ] Support service discovery with Kubernetes API

### v0.x.x

- [ ] TBD