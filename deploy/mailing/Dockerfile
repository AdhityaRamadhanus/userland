FROM golang:1.12.9

WORKDIR /go/src/github.com/AdhityaRamadhanus/userland
COPY . .

RUN make build-mail

EXPOSE 8001
ENTRYPOINT ["./mail"]