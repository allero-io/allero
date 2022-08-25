FROM golang:1.19-alpine AS builder
RUN apk --no-cache add curl gcc g++

WORKDIR /go/src/app
COPY . .

RUN curl -H "Authorization: token $GITHUB_TOKEN" --silent "https://api.github.com/repos/allero-io/allero-wip/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/' > cli-version
RUN go get -d -v ./...
RUN go build -tags main -ldflags="-X github.com/allero-io/allero/cmd.CliVersion=$(cat cli-version)" -v

FROM alpine:3.14
COPY --from=builder /go/src/app/allero /
ENTRYPOINT ["/allero"]