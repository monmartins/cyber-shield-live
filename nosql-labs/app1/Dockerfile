FROM golang:1.22-alpine AS builder
RUN apk add --no-cache git
WORKDIR /app
ENV GOPROXY=direct
ENV GONOSUMDB=*
COPY . . 
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o ciphernote .

FROM alpine:3.19
WORKDIR /app
COPY --from=builder /app/ciphernote .
COPY static  ./static
COPY templates ./templates
EXPOSE 8081
CMD ["./ciphernote"]
