FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY go.mod ./
COPY . .
RUN go mod tidy && CGO_ENABLED=0 go build -o /imperium-api .

FROM alpine:3.20
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /imperium-api .
COPY db/migrations/ db/migrations/

EXPOSE 8090
CMD ["./imperium-api"]
