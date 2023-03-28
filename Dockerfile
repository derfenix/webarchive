FROM golang:1.20-alpine as builder

WORKDIR /project
ADD go.* ./
RUN go mod download
ADD . .
RUN CGO_ENABLED=0 go build -o service ./cmd/service/main.go

FROM surnet/alpine-wkhtmltopdf:3.17.0-0.12.6-full

WORKDIR /project
COPY --from=builder /project/service service
ENTRYPOINT ["./service"]
