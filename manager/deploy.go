package manager

import (
	"io"

	"github.com/pkg/errors"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/release"
)

var Settings = cli.New()

//Deploy :
type Deploy struct {
	*action.Install
	NoDeps bool
}

func RunDeploy(args []string, client *Deploy, valueOpts *values.Options, out io.Writer) (*release.Release, error) {
	Debug("Original chart version: %q", client.Version)
	if client.Version == "" && client.Devel {
		Debug("setting version to >0.0.0-0")
		client.Version = ">0.0.0-0"
	}

	name, chart, err := client.NameAndChart(args)
	if err != nil {
		return nil, err
	}
	client.ReleaseName = name

	cp, err := client.ChartPathOptions.LocateChart(chart, Settings)
	if err != nil {
		return nil, err
	}

	Debug("CHART PATH: %s\n", cp)

	p := getter.All(Settings)
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		return nil, err
	}

	// Check chart dependencies to make sure all are present in /charts
	chartRequested, err := Load(cp, client)
	if err != nil {
		return nil, err
	}

	if err := checkIfInstallable(chartRequested); err != nil {
		return nil, err
	}

	if chartRequested.Metadata.Deprecated {
		Warning("This chart is deprecated")
	}
	if !client.NoDeps {
		if req := chartRequested.Metadata.Dependencies; req != nil {
			// If CheckDependencies returns an error, we have unfulfilled dependencies.
			// As of Helm 2.4.0, this is treated as a stopping condition:
			// https://github.com/helm/helm/issues/2209
			if err := action.CheckDependencies(chartRequested, req); err != nil {
				if client.DependencyUpdate {
					man := &downloader.Manager{
						Out:              out,
						ChartPath:        cp,
						Keyring:          client.ChartPathOptions.Keyring,
						SkipUpdate:       false,
						Getters:          p,
						RepositoryConfig: Settings.RepositoryConfig,
						RepositoryCache:  Settings.RepositoryCache,
						Debug:            Settings.Debug,
					}
					if err := man.Update(); err != nil {
						return nil, err
					}
					// Reload the chart with the updated Chart.lock file.
					if chartRequested, err = loader.Load(cp); err != nil {
						return nil, errors.Wrap(err, "failed reloading chart after repo update")
					}
				} else {
					return nil, err
				}
			}
		}
	}

	client.Namespace = Settings.Namespace()
	return client.Run(chartRequested, vals)
}

// checkIfInstallable validates if a chart can be installed
//
// Application chart type is only installable
func checkIfInstallable(ch *chart.Chart) error {
	switch ch.Metadata.Type {
	case "", "application":
		return nil
	}
	return errors.Errorf("%s charts are not installable", ch.Metadata.Type)
}
