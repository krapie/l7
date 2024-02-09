# Usage

## Basic Usage

```bash
# Start the load balancer
./bin/plumber
```

## Docker Usage

We use Docker Compose to test Plumber.

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

## Docker Usage with Yorkie

We use Docker Compose to test Plumber with Yorkie.

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

## Kubernetes Usage with Yorkie

We use Minikube to test Plumber with Yorkie.

```bash
# Start minikube with ingress addon
minikube start
minikube addons enable ingress

# Install Yorkie cluster with helm chart
helm install yorkie-cluster ./build/charts/yorkie-cluster

# Expose Yorkie cluster with Minikube tunnel
minikube tunnel

# Test with yorkie-js-sdk
git clone https://github.com/yorkie-team/yorkie-js-sdk.git
cd yorkie-js-sdk
npm install
sed -i html 's#http://localhost:8080#http://localhost#g' ./public/index.html
npm run dev

# Cleanup the minikube
minikube delete
```
