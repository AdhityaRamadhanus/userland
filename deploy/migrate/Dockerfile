FROM alpine:3.8

WORKDIR /migration
COPY pkg/storage/postgres/migration /migration

RUN apk add --no-cache curl && \
    curl -L https://github.com/golang-migrate/migrate/releases/download/v4.1.0/migrate.linux-amd64.tar.gz | tar xvz
ENTRYPOINT ["./migrate.linux-amd64", "-path", "."]