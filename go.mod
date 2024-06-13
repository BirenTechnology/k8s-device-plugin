module github.com/BirenTechnology/k8s-device-plugin

go 1.16

require (
	github.com/BirenTechnology/go-brml v0.0.0-20240612073547-7d6adadc1c0b
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/kubevirt/device-plugin-manager v1.19.4
	github.com/onsi/ginkgo v1.16.5 // indirect
	github.com/sirupsen/logrus v1.9.0
	github.com/spf13/cobra v1.7.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.8.3
	google.golang.org/grpc v1.56.3
	k8s.io/client-go v0.28.4
	k8s.io/kubelet v0.28.4
	sigs.k8s.io/yaml v1.3.0
	tags.cncf.io/container-device-interface/specs-go v0.6.0
)
