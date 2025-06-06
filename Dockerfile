############# builder
FROM golang:1.24.4 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-networking-calico

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG EFFECTIVE_VERSION

RUN make install EFFECTIVE_VERSION=$EFFECTIVE_VERSION

############# gardener-extension-networking-calico
FROM gcr.io/distroless/static-debian12:nonroot AS gardener-extension-networking-calico
WORKDIR /

COPY --from=builder /go/bin/gardener-extension-networking-calico /gardener-extension-networking-calico
ENTRYPOINT ["/gardener-extension-networking-calico"]

############# gardener-extension-admission-calico
FROM gcr.io/distroless/static-debian12:nonroot AS gardener-extension-admission-calico
WORKDIR /

COPY --from=builder /go/bin/gardener-extension-admission-calico /gardener-extension-admission-calico
ENTRYPOINT ["/gardener-extension-admission-calico"]
