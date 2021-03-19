module github.com/derailed/k9s

go 1.16

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
)

require (
	github.com/atotto/clipboard v0.1.4
	github.com/cenkalti/backoff v2.2.1+incompatible
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/derailed/popeye v0.9.0
	github.com/derailed/tview v0.4.10
	github.com/drone/envsubst v1.0.2 // indirect
	github.com/fatih/color v1.10.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/fvbommel/sortorder v1.0.2
	github.com/gdamore/tcell/v2 v2.2.0
	github.com/ghodss/yaml v1.0.0
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/mattn/go-runewidth v0.0.10
	github.com/openfaas/faas v0.0.0-20200207215241-6afae214e3ec
	github.com/openfaas/faas-cli v0.0.0-20200124160744-30b7cec9634c
	github.com/openfaas/faas-provider v0.15.0
	github.com/petergtz/pegomock v2.9.0+incompatible
	github.com/rakyll/hey v0.1.4
	github.com/rs/zerolog v1.20.0
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/sahilm/fuzzy v0.1.0
	github.com/spf13/cobra v1.1.3
	github.com/stretchr/testify v1.7.0
	golang.org/x/text v0.3.5
	gopkg.in/yaml.v2 v2.4.0
	helm.sh/helm/v3 v3.5.2
	k8s.io/api v0.20.5
	k8s.io/apimachinery v0.20.5
	k8s.io/cli-runtime v0.20.5
	k8s.io/client-go v0.20.5
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.4.0 // indirect
	k8s.io/kubectl v0.20.5
	k8s.io/metrics v0.20.5
	sigs.k8s.io/yaml v1.2.0
)
