package helmdeploy

import (
	"fmt"
	"io"
	"strings"
	"time"

	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli/output"
	"helm.sh/helm/v3/pkg/release"
)

type statusPrinter struct {
	release         *release.Release
	debug           bool
	showDescription bool
}

func (s statusPrinter) WriteJSON(out io.Writer) error {
	return output.EncodeJSON(out, s.release)
}

func (s statusPrinter) WriteYAML(out io.Writer) error {
	return output.EncodeYAML(out, s.release)
}

func (s statusPrinter) WriteTable(out io.Writer) error {
	if s.release == nil {
		return nil
	}
	fmt.Fprintf(out, "NAME: %s\n", s.release.Name)
	if !s.release.Info.LastDeployed.IsZero() {
		fmt.Fprintf(out, "LAST DEPLOYED: %s\n", s.release.Info.LastDeployed.Format(time.ANSIC))
	}
	fmt.Fprintf(out, "NAMESPACE: %s\n", s.release.Namespace)
	fmt.Fprintf(out, "STATUS: %s\n", s.release.Info.Status.String())
	fmt.Fprintf(out, "REVISION: %d\n", s.release.Version)
	if s.showDescription {
		fmt.Fprintf(out, "DESCRIPTION: %s\n", s.release.Info.Description)
	}

	executions := executionsByHookEvent(s.release)
	if tests, ok := executions[release.HookTest]; !ok || len(tests) == 0 {
		fmt.Fprintln(out, "TEST SUITE: None")
	} else {
		for _, h := range tests {
			// Don't print anything if hook has not been initiated
			if h.LastRun.StartedAt.IsZero() {
				continue
			}
			fmt.Fprintf(out, "TEST SUITE:     %s\n%s\n%s\n%s\n",
				h.Name,
				fmt.Sprintf("Last Started:   %s", h.LastRun.StartedAt.Format(time.ANSIC)),
				fmt.Sprintf("Last Completed: %s", h.LastRun.CompletedAt.Format(time.ANSIC)),
				fmt.Sprintf("Phase:          %s", h.LastRun.Phase),
			)
		}
	}

	if s.debug {
		fmt.Fprintln(out, "USER-SUPPLIED VALUES:")
		err := output.EncodeYAML(out, s.release.Config)
		if err != nil {
			return err
		}
		// Print an extra newline
		fmt.Fprintln(out)

		cfg, err := chartutil.CoalesceValues(s.release.Chart, s.release.Config)
		if err != nil {
			return err
		}

		fmt.Fprintln(out, "COMPUTED VALUES:")
		err = output.EncodeYAML(out, cfg.AsMap())
		if err != nil {
			return err
		}
		// Print an extra newline
		fmt.Fprintln(out)
	}

	if strings.EqualFold(s.release.Info.Description, "Dry run complete") || s.debug {
		fmt.Fprintln(out, "HOOKS:")
		for _, h := range s.release.Hooks {
			fmt.Fprintf(out, "---\n# Source: %s\n%s\n", h.Path, h.Manifest)
		}
		fmt.Fprintf(out, "MANIFEST:\n%s\n", s.release.Manifest)
	}

	if len(s.release.Info.Notes) > 0 {
		fmt.Fprintf(out, "NOTES:\n%s\n", strings.TrimSpace(s.release.Info.Notes))
	}
	return nil
}

func executionsByHookEvent(rel *release.Release) map[release.HookEvent][]*release.Hook {
	result := make(map[release.HookEvent][]*release.Hook)
	for _, h := range rel.Hooks {
		for _, e := range h.Events {
			executions, ok := result[e]
			if !ok {
				executions = []*release.Hook{}
			}
			result[e] = append(executions, h)
		}
	}
	return result
}
