## Build first
FROM golang:latest as builder
RUN mkdir /builddir
ADD go.mod go.sum /builddir/
ADD cmd /builddir/cmd
ADD config /builddir/config
ADD crypto /builddir/crypto
ADD model /builddir/model
ADD bot /builddir/bot
WORKDIR /builddir
RUN go mod download
RUN go mod tidy
RUN go mod verify
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-w -s -extldflags "-static"' -o arrgo \
    github.com/wneessen/arrgo/cmd/arrgo

## Create scratch image
FROM scratch
LABEL maintainer="wn@neessen.net"
COPY ["docker-files/passwd", "/etc/passwd"]
COPY ["docker-files/group", "/etc/group"]
COPY ["arrgo.toml.example", "/arrgo/etc/arrgo.toml"]
COPY --from=builder ["/etc/ssl/certs/ca-certificates.crt", "/etc/ssl/cert.pem"]
COPY --chown=arrgo ["LICENSE", "/arrgo/LICENSE"]
COPY --chown=arrgo ["README.md", "/arrgo/README.md"]
COPY --from=builder --chown=arrgo ["/builddir/arrgo", "/arrgo/arrgo"]
ADD --chown=arrgo sql_migrations /arrgo/sql_migrations
WORKDIR /arrgo
USER arrgo
VOLUME ["/arrgo/etc"]
ENTRYPOINT ["/arrgo/arrgo"]