FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o orchestrator .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates docker-cli

WORKDIR /app
COPY --from=builder /app/orchestrator .

EXPOSE 8080
CMD ["./orchestrator"]
