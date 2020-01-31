############# builder
FROM golang:1.13.4 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-networking-calico
COPY . .
RUN make install-requirements && make VERIFY=true all

############# gardener-extension-networking-calico
FROM alpine:3.11.3 AS gardener-extension-networking-calico

COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-networking-calico /gardener-extension-networking-calico
ENTRYPOINT ["/gardener-extension-networking-calico"]
