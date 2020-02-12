FROM golang:1.13 AS build

WORKDIR /go/src/pastewatch

COPY . .

RUN GOOS=linux CGO_ENABLED=0 go build -a -o pastewatch .

FROM alpine

WORKDIR /
COPY --from=build /go/src/pastewatch/pastewatch /usr/local/bin/pastewatch

EXPOSE 8080

RUN adduser -D -g '' pastewatch
USER pastewatch

ENTRYPOINT ["pastewatch"]
