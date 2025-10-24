FROM golang:1.25.1-alpine3.21@sha256:331bde41663c297cba0f5abf37e929be644f3cbd84bf45f49b0df9d774f4d912 AS builder

WORKDIR $GOPATH/src/github.com/sidey79/scantopl/

COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/scantopl

FROM alpine:3.22.2@sha256:4b7ce07002c69e8f3d704a9c5d6fd3053be500b7f1c69fc0d80990c2ad8dd412
COPY --from=builder /go/bin/scantopl /usr/bin/scantopl

ENV \
  # The paperless instance URL
  PLURL="http://127.0.0.1:8080" \
  # The paperless token
  PLTOKEN="XXXXXXXXXXXXXXXXXXXXXXX"

ENTRYPOINT ["/usr/bin/scantopl", "-scandir", "/output"]
