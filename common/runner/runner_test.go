package runner

import (
	"context"
	"os"
	"testing"

	"github.com/docktermj/go-logger/logger"
)

/*
 * The unit tests in this file simulate command line invocation.
 */

func mockFunction(ctx context.Context, argv []string) {
	logger.Debug(">>>> Success")
}

func TestRun(test *testing.T) {

	os.Args = []string{"bixserver", "configuration", "--debug"}
	args := os.Args[1:]
	logger.Debugf("%+v", args)
	usage := `
Usage:
    bixserver configuration [<args>...]

Subcommands:
    configuration    View configuration
`

	functions := map[string]interface{}{
		"configuration": mockFunction,
	}

	Run(context.Background(), args, functions, usage)
}
