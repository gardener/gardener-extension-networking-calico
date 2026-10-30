############# builder
FROM golang:1.26.10 AS builder

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
FROM golang:1.26.10 AS cni-plugins-builder
ARG TARGETARCH
ARG CNI_PLUGINS_VERSION=v1.9.1
ARG CNI_PLUGINS_SHA256_AMD64=b98f74a0f8522f0a83867178729c1aa70f2158f90c45a2ca8fa791db1c76b303
ARG CNI_PLUGINS_SHA256_ARM64=56171987d3947707c3563db2f4001bccaf50fd63468611b9f3cbecb1375ee7ec
WORKDIR /
RUN mkdir -p /usr/src/cni/bin && \
    case "${TARGETARCH}" in \
      amd64) CNI_PLUGINS_SHA256="${CNI_PLUGINS_SHA256_AMD64}" ;; \
      arm64) CNI_PLUGINS_SHA256="${CNI_PLUGINS_SHA256_ARM64}" ;; \
      *) echo "unsupported architecture: ${TARGETARCH}" && exit 1 ;; \
    esac && \
    curl -L -O "https://github.com/containernetworking/plugins/releases/download/${CNI_PLUGINS_VERSION}/cni-plugins-linux-${TARGETARCH}-${CNI_PLUGINS_VERSION}.tgz" && \
    echo "${CNI_PLUGINS_SHA256}  cni-plugins-linux-${TARGETARCH}-${CNI_PLUGINS_VERSION}.tgz" | sha256sum -c && \
    tar -xvf "cni-plugins-linux-${TARGETARCH}-${CNI_PLUGINS_VERSION}.tgz" -C /usr/src/cni/bin/ && \
    echo done
############# cni-plugins
FROM alpine:3.24.3 AS cni-plugins
WORKDIR /
LABEL io.k8s.display-name="Container Network Plugins"

COPY --from=cni-plugins-builder /usr/src/cni/bin/* /usr/src/cni/bin/
