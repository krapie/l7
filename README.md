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
# Start the load balancer with 2 backends
./bin/plumber --backends http://localhost:8080,http://localhost:8081
```

Docker-compose testing:

```bash
# Test the load balancer with docker-compose
make docker-compose-up

# Start the load balancer
./bin/plumber

# Send a request to the load balancer
chmod +x ./scripts/lb_distribution_test.sh
./scripts/lb_distribution_test.sh
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
- [ ] Support mechanism to resolve split-brain of long-lived connection

### v0.3.0

- [ ] Support interceptor to modify request/response
- [ ] Support service discovery with Kubernetes API

### v0.x.x

- [ ] Support rate limiter
- [ ] TBD
