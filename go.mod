module github.com/thynquest/helm-deploy

go 1.13

require (
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.0.0
	github.com/spf13/pflag v1.0.5
	helm.sh/helm/v3 v3.3.4
	k8s.io/helm v2.16.12+incompatible
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/yaml v1.2.0
)
