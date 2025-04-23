FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o app cmd/bff/main.go

FROM golang:1.24-alpine AS runner
WORKDIR /app
COPY --from=builder /app/cert ./cert
COPY --from=builder /app/app ./
EXPOSE 443
CMD ["./app"]