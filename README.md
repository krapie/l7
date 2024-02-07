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

```bash
# Start the load balancer with 2 backends
./bin/plumber --backends http://localhost:8080,http://localhost:8081
```

## Roadmap

Plumber aims to support [Yorkie](https://github.com/yorkie-team/yorkie) as a backend for the load balancer.
The following features are planned to be implemented first:

### v0.1.0

- [x] Support static load balancing with round-robin algorithm
- [x] Support backends health check 

### v0.2.0

- [x] Support consistent hashing algorithm (maglev with siphash)
- [ ] Support dynamic backend configuration (with K8s API)
- [ ] Support mechanism to resolve split-brain of long-lived connection

### v0.3.0

- [ ] Support interceptor to modify request/response
- [ ] TBD
