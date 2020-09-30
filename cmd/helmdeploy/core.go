package helmdeploy

import (
	"fmt"
	"os"

	"helm.sh/helm/v3/pkg/cli"
)

var settings = cli.New()

func warning(format string, v ...interface{}) {
	format = fmt.Sprintf("WARNING: %s\n", format)
	fmt.Fprintf(os.Stderr, format, v...)
}
