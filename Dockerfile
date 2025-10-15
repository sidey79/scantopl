FROM golang:1.25.3-alpine3.21@sha256:0c9f3e09a50a6c11714dbc37a6134fd0c474690030ed07d23a61755afd3a812f AS builder

WORKDIR $GOPATH/src/github.com/sidey79/scantopl/

COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/scantopl

FROM alpine:3.22.1@sha256:4bcff63911fcb4448bd4fdacec207030997caf25e9bea4045fa6c8c44de311d1
COPY --from=builder /go/bin/scantopl /usr/bin/scantopl

ENV \
  # The paperless instance URL
  PLURL="http://127.0.0.1:8080" \
  # The paperless token
  PLTOKEN="XXXXXXXXXXXXXXXXXXXXXXX"

ENTRYPOINT ["/usr/bin/scantopl", "-scandir", "/output"]
