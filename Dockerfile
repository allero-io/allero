FROM golang:1.19-alpine AS builder
RUN apk --no-cache add curl gcc g++

ARG cli-version

WORKDIR /go/src/app
COPY . .

RUN go get -d -v ./...
RUN go build -tags main -ldflags="-X github.com/allero-io/allero/cmd.CliVersion=${cli-version}" -v

FROM alpine:3.14
COPY --from=builder /go/src/app/allero /
ENTRYPOINT ["/allero"]
