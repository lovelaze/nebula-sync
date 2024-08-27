FROM golang:1.23-alpine AS golang

WORKDIR /app

RUN apk add -U tzdata upx && \
    apk --update add ca-certificates

COPY . .

RUN go mod download
RUN go mod verify

ARG VERSION=dev

ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOFLAGS="-a -trimpath -ldflags=-w -ldflags=-s -ldflags=-X=github.com/lovelaze/nebula-sync/version.Version=${VERSION} -o=nebula-sync"

RUN go build . && \
    upx -q nebula-sync

FROM scratch

COPY --link --from=golang /usr/share/zoneinfo/ /usr/share/zoneinfo/
COPY --link --from=golang /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --link --from=golang /app/nebula-sync /usr/local/bin/

USER 1001

ENTRYPOINT ["nebula-sync"]
CMD ["run"]
