FROM golang:1.25-alpine3.21 AS builder
RUN apk --no-cache add bash make gcc g++
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o /server ./cmd/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates postgresql-client
WORKDIR /app
COPY --from=builder /server ./
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./server"]
