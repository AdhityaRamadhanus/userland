FROM golang:1.12.9

WORKDIR /go/src/github.com/AdhityaRamadhanus/userland
COPY . .

RUN make build-api

EXPOSE 8000
ENTRYPOINT ["./api"]