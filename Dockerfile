############# builder
FROM golang:1.17.6 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-networking-calico
COPY . .
RUN make install

############# gardener-extension-networking-calico
FROM alpine:3.13.7 AS gardener-extension-networking-calico

COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-networking-calico /gardener-extension-networking-calico
ENTRYPOINT ["/gardener-extension-networking-calico"]

############# gardener-extension-admission-calico
FROM alpine:3.13.7 AS gardener-extension-admission-calico

COPY --from=builder /go/bin/gardener-extension-admission-calico /gardener-extension-admission-calico
ENTRYPOINT ["/gardener-extension-admission-calico"]
