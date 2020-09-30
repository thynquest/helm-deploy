package main

import (
	"os"

	"github.com/thynquest/helm-deploy/cmd/helmdeploy"
	"helm.sh/helm/v3/pkg/action"
)

func main() {
	actionConfig := new(action.Configuration)
	cmd := helmdeploy.NewDeployCmd(actionConfig, os.Stdout)
}
