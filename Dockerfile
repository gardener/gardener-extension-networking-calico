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
ARG CNI_PLUGINS_VERSION=v1.10.0
ARG CNI_PLUGINS_SHA256_AMD64=ce8603d6a1bd448efba5bd6dc44f38dce349f821aa57e1bb354b52628ebf1853
ARG CNI_PLUGINS_SHA256_ARM64=4f25d7f136be368746185fba6557d799cf3e1890f7556d450744ea20d300d68c
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
FROM alpine:3.24.4 AS cni-plugins
WORKDIR /
LABEL io.k8s.display-name="Container Network Plugins"

COPY --from=cni-plugins-builder /usr/src/cni/bin/* /usr/src/cni/bin/
