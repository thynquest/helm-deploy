package helmdeploy

import (
	"io"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/thynquest/helm-deploy/manager"
	"helm.sh/helm/v3/cmd/helm/require"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli/output"
	"helm.sh/helm/v3/pkg/cli/values"
)

const deployDesc = `
This command install a chart archive with the possibility to skip the dependency during the install process.
`

//NewDeploy : new deploy operation
func NewDeploy(config *action.Configuration) *manager.Deploy {
	return &manager.Deploy{
		Install: action.NewInstall(config),
	}
}

//NewDeployCmd :
func NewDeployCmd(cfg *action.Configuration, out io.Writer) *cobra.Command {
	client := NewDeploy(cfg)
	valueOpts := &values.Options{}
	var outfmt output.Format

	cmd := &cobra.Command{
		Use:   "deploy [NAME] [CHART]",
		Short: "deploy a chart",
		Long:  deployDesc,
		Args:  require.MinimumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			rel, err := manager.RunDeploy(args, client, valueOpts, out)
			if err != nil {
				return err
			}

			return outfmt.Write(out, &statusPrinter{rel, manager.Settings.Debug, false})
		},
	}

	addDeployFlags(cmd, cmd.Flags(), client, valueOpts)
	return cmd
}

func addDeployFlags(cmd *cobra.Command, f *pflag.FlagSet, client *manager.Deploy, valueOpts *values.Options) {
	f.BoolVar(&client.CreateNamespace, "create-namespace", false, "create the release namespace if not present")
	f.BoolVar(&client.DryRun, "dry-run", false, "simulate an install")
	f.BoolVar(&client.DisableHooks, "no-hooks", false, "prevent hooks from running during install")
	f.BoolVar(&client.Replace, "replace", false, "re-use the given name, only if that name is a deleted release which remains in the history. This is unsafe in production")
	f.DurationVar(&client.Timeout, "timeout", 300*time.Second, "time to wait for any individual Kubernetes operation (like Jobs for hooks)")
	f.BoolVar(&client.Wait, "wait", false, "if set, will wait until all Pods, PVCs, Services, and minimum number of Pods of a Deployment, StatefulSet, or ReplicaSet are in a ready state before marking the release as successful. It will wait for as long as --timeout")
	f.BoolVarP(&client.GenerateName, "generate-name", "g", false, "generate the name (and omit the NAME parameter)")
	f.StringVar(&client.NameTemplate, "name-template", "", "specify template used to name the release")
	f.StringVar(&client.Description, "description", "", "add a custom description")
	f.BoolVar(&client.Devel, "devel", false, "use development versions, too. Equivalent to version '>0.0.0-0'. If --version is set, this is ignored")
	f.BoolVar(&client.DependencyUpdate, "dependency-update", false, "run helm dependency update before installing the chart")
	f.BoolVar(&client.DisableOpenAPIValidation, "disable-openapi-validation", false, "if set, the installation process will not validate rendered templates against the Kubernetes OpenAPI Schema")
	f.BoolVar(&client.Atomic, "atomic", false, "if set, the installation process deletes the installation on failure. The --wait flag will be set automatically if --atomic is used")
	f.BoolVar(&client.SkipCRDs, "skip-crds", false, "if set, no CRDs will be installed. By default, CRDs are installed if not already present")
	f.BoolVar(&client.NoDeps, "no-deps", false, "if set, no dependencies will be installed. By default, dependencies are installed ")
	f.BoolVar(&client.SubNotes, "render-subchart-notes", false, "if set, render subchart notes along with the parent")
}
