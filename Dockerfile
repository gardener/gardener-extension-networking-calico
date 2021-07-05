############# builder
FROM eu.gcr.io/gardener-project/3rd/golang:1.16.2 AS builder

WORKDIR /go/src/github.com/gardener/gardener-extension-networking-calico
COPY . .
RUN make install

############# gardener-extension-networking-calico
FROM eu.gcr.io/gardener-project/3rd/alpine:3.13.4 AS gardener-extension-networking-calico

COPY charts /charts
COPY --from=builder /go/bin/gardener-extension-networking-calico /gardener-extension-networking-calico
ENTRYPOINT ["/gardener-extension-networking-calico"]

############# gardener-extension-admission-calico
FROM eu.gcr.io/gardener-project/3rd/alpine:3.13.4 AS gardener-extension-admission-calico

COPY --from=builder /go/bin/gardener-extension-admission-calico /gardener-extension-admission-calico
ENTRYPOINT ["/gardener-extension-admission-calico"]
