ARG BUILD_IMAGE=golang:1.22-alpine3.19
ARG BASE_IMAGE=alpine-3.19
ARG BASE_DISTROLESS_IMAGE=gcr.io/distroless/static-debian12

FROM --platform=${TARGETPLATFORM} ${BUILD_IMAGE} as builder
ARG TARGETARCH
ENV ARCH=${TARGETARCH}
RUN apk add --update --no-cache ca-certificates bash make gcc musl-dev git openssh wget curl

WORKDIR /go/src/bifrost

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 make build ARCH=${ARCH}

#####

FROM --platform=${TARGETPLATFORM} ${BASE_DISTROLESS_IMAGE}

COPY --from=builder /go/src/bifrost/build/_output/bin/bifrost /bifrost
COPY --from=builder /go/src/bifrost/build/config/config.json config/config.json

ENTRYPOINT [ "/bifrost" ]
