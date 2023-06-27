ARG BUILD_IMAGE=golang:1.20.5@sha256:fd8d0f8f05c0254d80cfb040e2c6351e477593b7dbf24b0d495ba1e97aa14146
ARG BASE_IMAGE=gcr.io/distroless/static-debian11@sha256:5759d194607e472ff80fff5833442d3991dd89b219c96552837a2c8f74058617

FROM ${BUILD_IMAGE} as builder

RUN apk add --update --no-cache ca-certificates bash make gcc musl-dev git openssh wget curl

WORKDIR /go/src/bifrost

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 make build

#####

FROM ${BASE_IMAGE}

COPY --from=builder /go/src/bifrost/bifrost /bifrost
COPY --from=builder /go/src/bifrost/build/config/config.json config/config.json

ENTRYPOINT [ "/bifrost" ]
