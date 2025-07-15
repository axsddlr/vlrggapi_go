# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o vlrggapi ./cmd

# Final image
FROM alpine:3.20

WORKDIR /app

COPY --from=builder /app/vlrggapi .
COPY --from=builder /app/internal ./internal
COPY --from=builder /app/go.mod .
COPY --from=builder /app/go.sum .

EXPOSE 3001

ENV PORT=3001

CMD ["./vlrggapi"]
