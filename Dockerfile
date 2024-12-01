FROM golang:1.23.3-alpine3.20 AS builder
LABEL authors="usman243"
WORKDIR /home/usr/app
RUN apk --no-cache add make
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN make build/docker

FROM scratch as final
ENV HOME=/home/usr/app/bin
WORKDIR $HOME
COPY --from=builder $HOME/letschat-api $HOME
RUN addgroup -S letschatGroup && adduser -S letschat -G letschatGroup
USER letschat:letschatGroup
ENTRYPOINT ["./letschat", "-db-dsn=${LETSCHAT_API_DB_DSN}"]
