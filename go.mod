module github.com/derailed/k9s

go 1.13

replace helm.sh/helm/v3 => /Users/fernand/go_wk/derailed/src/github.com/derailed/tmp/helm

require (
	github.com/atotto/clipboard v0.1.2
	github.com/derailed/popeye v0.8.1
	github.com/derailed/tview v0.3.10
	github.com/drone/envsubst v1.0.2 // indirect
	github.com/fatih/color v1.9.0
	github.com/fsnotify/fsnotify v1.4.7
	github.com/gdamore/tcell v1.3.0
	github.com/ghodss/yaml v1.0.0
	github.com/kylelemons/godebug v1.1.0 // indirect
	github.com/mattn/go-runewidth v0.0.9
	github.com/openfaas/faas v0.0.0-20200207215241-6afae214e3ec
	github.com/openfaas/faas-cli v0.0.0-20200124160744-30b7cec9634c
	github.com/openfaas/faas-provider v0.15.0
	github.com/petergtz/pegomock v2.7.0+incompatible
	github.com/rakyll/hey v0.1.3
	github.com/rs/zerolog v1.18.0
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/sahilm/fuzzy v0.1.0
	github.com/spf13/cobra v0.0.6
	github.com/stretchr/testify v1.5.1
	golang.org/x/text v0.3.2
	gopkg.in/yaml.v2 v2.2.8
	helm.sh/helm/v3 v3.1.2
	k8s.io/api v0.18.0
	k8s.io/apimachinery v0.18.0
	k8s.io/cli-runtime v0.18.0
	k8s.io/client-go v0.18.0
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.18.0
	k8s.io/metrics v0.18.0
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/yaml v1.2.0
	vbom.ml/util v0.0.0-20180919145318-efcd4e0f9787
)
