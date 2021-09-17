module github.com/gardener/machine-controller-manager-provider-gcp

go 1.15

require (
	github.com/gardener/machine-controller-manager v0.40.1-0.20210916131306-0b3c2244f009
	github.com/onsi/ginkgo v1.16.2
	github.com/onsi/gomega v1.11.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/pflag v1.0.5
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b
	golang.org/x/net v0.0.0-20210326060303-6b1517762897
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	google.golang.org/api v0.20.0
	k8s.io/api v0.20.6
	k8s.io/apimachinery v0.20.6
	k8s.io/component-base v0.20.6
	k8s.io/klog v1.0.0
	sigs.k8s.io/yaml v1.2.0
)
