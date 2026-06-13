FROM golang:1.24-alpine AS builder
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w" -trimpath -o /api ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=builder /api /api
EXPOSE 8080
ENTRYPOINT ["/api"]
