# Build go binary
FROM golang:1.21 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN make build

# Build docker image with go binary with certs
FROM alpine:latest

RUN apk --no-cache add ca-certificates

COPY --from=builder /app/bin/plumber /bin/

EXPOSE 80 80

ENTRYPOINT ["plumber"]
