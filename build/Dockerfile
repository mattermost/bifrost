
FROM golang:1.17.3-alpine3.14 AS builder

RUN apk add --update --no-cache ca-certificates bash make gcc musl-dev git openssh wget curl

WORKDIR /go/src/bifrost

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 make build

#####

FROM alpine:3.14

COPY --from=builder /go/src/bifrost/build/_output/bin/bifrost /bifrost
COPY --from=builder /go/src/bifrost/build/config/config.json config/config.json

ENTRYPOINT [ "/bifrost" ]
