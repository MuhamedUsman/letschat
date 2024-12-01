FROM golang:1.23.3-alpine3.20 AS builder
LABEL authors="usman243"
WORKDIR /home/usr/app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags='-s -w' -o ./bin/letschat-api ./cmd/letschat-api

FROM scratch as final
ENV HOME=/home/usr/app/bin
WORKDIR $HOME
COPY --from=builder $HOME/letschat-api $HOME
RUN addgroup -S letschatGroup && adduser -S letschat -G letschatGroup
USER letschat:letschatGroup
ENTRYPOINT ["./letschat", "-db-dsn=${LETSCHAT_API_DB_DSN}"]
