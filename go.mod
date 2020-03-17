module github.com/deislabs/smi-metrics

go 1.12

require (
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.4.2 // indirect
	github.com/deislabs/smi-sdk-go v0.1.0
	github.com/eknkc/amber v0.0.0-20171010120322-cdade1c07385 // indirect
	github.com/fsnotify/fsnotify v1.4.7
	github.com/go-chi/chi v4.0.2+incompatible
	github.com/grpc-ecosystem/go-grpc-middleware v1.0.0 // indirect
	github.com/hellofresh/health-go v2.0.2+incompatible
	github.com/huandu/xstrings v1.2.0 // indirect
	github.com/imdario/mergo v0.3.7 // indirect
	github.com/improbable-eng/go-httpwares v0.0.0-20190118142334-33c6690a604c
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/masterminds/sprig v2.18.0+incompatible
	github.com/prometheus/client_golang v0.9.3-0.20190127221311-3c4408c8b829
	github.com/prometheus/common v0.3.0
	github.com/sirupsen/logrus v1.4.1
	github.com/spf13/cobra v0.0.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.3.2
	github.com/stretchr/testify v1.4.0
	github.com/unrolled/render v1.0.0
	google.golang.org/grpc v1.20.1 // indirect
	gopkg.in/yaml.v2 v2.2.4
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
	k8s.io/klog v1.0.0
)

replace github.com/deislabs/smi-sdk-go => github.com/deislabs/smi-sdk-go v0.0.0-20200313173708-210bdedfca08
