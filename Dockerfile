############# builder
FROM golang:1.25.5 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-networking-calico

# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG EFFECTIVE_VERSION

RUN make install EFFECTIVE_VERSION=$EFFECTIVE_VERSION

############# gardener-extension-networking-calico
FROM gcr.io/distroless/static-debian13:nonroot AS gardener-extension-networking-calico
WORKDIR /

COPY --from=builder /go/bin/gardener-extension-networking-calico /gardener-extension-networking-calico
ENTRYPOINT ["/gardener-extension-networking-calico"]

############# gardener-extension-admission-calico
FROM gcr.io/distroless/static-debian13:nonroot AS gardener-extension-admission-calico
WORKDIR /

COPY --from=builder /go/bin/gardener-extension-admission-calico /gardener-extension-admission-calico
ENTRYPOINT ["/gardener-extension-admission-calico"]

############# cni-plugins-builder
FROM golang:1.25.5 AS cni-plugins-builder
ARG CNI_PLUGINS_VERSION=v1.9.0
WORKDIR /
RUN mkdir -p /usr/src/cni/bin && \
    curl -L -O https://github.com/containernetworking/plugins/releases/download/${CNI_PLUGINS_VERSION}/cni-plugins-linux-amd64-${CNI_PLUGINS_VERSION}.tgz && \
    tar -xvf cni-plugins-linux-amd64-${CNI_PLUGINS_VERSION}.tgz -C /usr/src/cni/bin/ && \
    echo done
############# cni-plugins
FROM alpine:3.23.2 AS cni-plugins
WORKDIR /
LABEL io.k8s.display-name="Container Network Plugins"

COPY --from=cni-plugins-builder /usr/src/cni/bin/* /usr/src/cni/bin/
