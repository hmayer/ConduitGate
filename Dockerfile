# syntax=docker/dockerfile:1

FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o conduitgate .

FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY --from=builder /app/conduitgate /app/conduitgate
COPY routes.sample.json /app/routes.json

EXPOSE 8080

ENTRYPOINT ["/app/conduitgate"]
