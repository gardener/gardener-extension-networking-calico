module github.com/gardener/gardener-extension-networking-calico

go 1.13

require (
	github.com/ahmetb/gen-crd-api-reference-docs v0.1.5
	github.com/gardener/gardener v1.1.1-0.20200306090858-1aa0c7eb14d1
	github.com/gardener/gardener-extensions v1.4.1-0.20200306095919-e316500c1379
	github.com/gardener/gardener-resource-manager v0.10.0
	github.com/go-logr/logr v0.1.0
	github.com/gobuffalo/envy v1.9.0 // indirect
	github.com/gobuffalo/genny v0.6.0 // indirect
	github.com/gobuffalo/gogen v0.2.0 // indirect
	github.com/gobuffalo/mapi v1.2.1 // indirect
	github.com/gobuffalo/packr v1.30.1 // indirect
	github.com/gobuffalo/packr/v2 v2.8.0
	github.com/gobuffalo/syncx v0.1.0 // indirect
	github.com/golang/mock v1.3.1
	github.com/karrick/godirwalk v1.15.5 // indirect
	github.com/onsi/ginkgo v1.10.3
	github.com/onsi/gomega v1.7.1
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v0.0.6
	github.com/spf13/pflag v1.0.5
	golang.org/x/crypto v0.0.0-20200311171314-f7b00557c8c4 // indirect
	golang.org/x/sys v0.0.0-20200302150141-5c8b2ff67527 // indirect
	golang.org/x/tools v0.0.0-20200313205530-4303120df7d8 // indirect
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/code-generator v0.17.0
	k8s.io/component-base v0.17.0
	k8s.io/helm v2.16.1+incompatible
	sigs.k8s.io/controller-runtime v0.4.0
)

replace (
	k8s.io/api => k8s.io/api v0.0.0-20190918155943-95b840bb6a1f // kubernetes-1.16.0
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190918161926-8f644eb6e783 // kubernetes-1.16.0
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190913080033-27d36303b655 // kubernetes-1.16.0
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190918160949-bfa5e2e684ad // kubernetes-1.16.0
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190918160344-1fbdaa4c8d90 // kubernetes-1.16.0
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190912054826-cd179ad6a269 // kubernetes-1.16.0
	k8s.io/component-base => k8s.io/component-base v0.0.0-20190918160511-547f6c5d7090 // kubernetes-1.16.0
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190918161219-8c8f079fddc3 // kubernetes-1.16.0
)
