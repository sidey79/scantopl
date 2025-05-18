FROM golang:1.24.3 AS builder

WORKDIR $GOPATH/src/github.com/StarkZarn/scantopl/

COPY . .

RUN go get -d -v
RUN go build -o /go/bin/scantopl

FROM alpine
COPY --from=builder /go/bin/scantopl /usr/bin/scantopl

ENV \
  # The paperless instance URL
  PLURL="http://127.0.0.1:8080" \
  # The paperless token
  PLTOKEN="XXXXXXXXXXXXXXXXXXXXXXX"

ENTRYPOINT ["/usr/bin/scantopl", "-scandir", "/output"]
