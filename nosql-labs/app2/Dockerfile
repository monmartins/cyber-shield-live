FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
ENV GOPROXY=direct
ENV GONOSUMDB=*
COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o peopledir .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/peopledir .
COPY static  ./static
COPY templates ./templates
EXPOSE 8082
CMD ["./peopledir"]
