FROM golang:1.23.3-alpine3.20 AS builder
LABEL authors="usman"
WORKDIR /home/usr/app
RUN apk --no-cache add make
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN make build/docker

FROM alpine:3.20
ENV HOME=/home/usr/app/bin
WORKDIR $HOME
COPY --from=builder $HOME/letschat-api $HOME
RUN addgroup -S letschatGroup && adduser -S letschat -G letschatGroup
USER letschat:letschatGroup
ENTRYPOINT ["./letschat", "-db-dsn=${LETSCHAT_API_DB_DSN}"]
