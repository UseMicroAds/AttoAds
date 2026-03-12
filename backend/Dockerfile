FROM golang:1.22-alpine AS builder

RUN apk add --no-cache gcc musl-dev

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /bin/api ./cmd/api
RUN CGO_ENABLED=0 go build -o /bin/discovery ./cmd/discovery
RUN CGO_ENABLED=0 go build -o /bin/verifier ./cmd/verifier

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /bin/api /bin/api
COPY --from=builder /bin/discovery /bin/discovery
COPY --from=builder /bin/verifier /bin/verifier
COPY --from=builder /app/migrations /migrations

EXPOSE 8080
CMD ["/bin/api"]
