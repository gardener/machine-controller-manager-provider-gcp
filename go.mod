module github.com/gardener/machine-controller-manager-provider-gcp

go 1.15

require (
	github.com/gardener/machine-controller-manager v0.39.0
	github.com/onsi/ginkgo v1.15.2
	github.com/onsi/gomega v1.11.0
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.5.1 // indirect
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.5.1 // indirect
	golang.org/x/lint v0.0.0-20191125180803-fdd1cda4f05f
	golang.org/x/net v0.0.0-20201202161906-c7110b5ffcbb
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	google.golang.org/api v0.4.0
	k8s.io/api v0.16.8
	k8s.io/apimachinery v0.16.8
	k8s.io/component-base v0.16.8
	k8s.io/klog v1.0.0
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/prometheus/client_golang => github.com/prometheus/client_golang v0.9.2
	k8s.io/api => k8s.io/api v0.16.8 // v0.16.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.8 // v0.16.8
	k8s.io/apiserver => k8s.io/apiserver v0.16.8 // v0.16.8
	k8s.io/client-go => k8s.io/client-go v0.16.8 // v0.16.8
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.16.8 // v0.16.8
	k8s.io/code-generator => k8s.io/code-generator v0.16.8 // v0.16.8
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf
)
