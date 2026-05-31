FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o newshub .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/newshub .
COPY static ./static
COPY templates ./templates
EXPOSE 8083
CMD ["./newshub"]
