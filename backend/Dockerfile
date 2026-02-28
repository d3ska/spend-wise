FROM golang:1.26-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 go build -o /app ./cmd/api

FROM alpine:3.21

COPY --from=builder /app /app

EXPOSE 8080

ENTRYPOINT ["/app"]
