FROM golang:1.18.2-alpine3.15 AS pre

COPY . /go/project

WORKDIR /go/project

RUN go build ./cmd/logservice

FROM alpine:latest

COPY --from=pre /go/project/logservice .

CMD ["./logservice"]
