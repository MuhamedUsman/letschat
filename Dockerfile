FROM golang:1.23.3-alpine3.20 AS builder
LABEL authors="usman243"
WORKDIR /home/usr/app/bin
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o letschat-api ./cmd/letschat-api

FROM alpine:3.20
RUN apk --no-cache add ca-certificates

ENV HOME=/home/usr/app/bin
WORKDIR $HOME
RUN addgroup -S letschatGroup && adduser -S letschat -G letschatGroup
COPY --from=builder /home/usr/app/bin/letschat-api ./letschat-api
RUN chown letschat:letschatGroup letschat-api
USER letschat:letschatGroup
ENTRYPOINT ["./letschat-api"]
CMD ["-db-dsn=${LETSCHAT_API_DB_DSN}"]