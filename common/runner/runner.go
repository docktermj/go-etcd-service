// Run subcommands.

package runner

// Reference: http://stackoverflow.com/questions/6769020/go-map-of-functions

import (
	"context"

	"github.com/docktermj/go-etcd-service/common/help"
)

func Run(ctx context.Context, argv []string, functions map[string]interface{}, usage string) {

	// If no arguments something is wrong, so show help.

	if len(argv) == 0 {
		help.ShowHelp(usage)
	}

	// Parse command and arguments.

	cmd := argv[0]
	args := []string{}
	if len(argv) > 1 {
		args = argv[1:]
	}

	// Call function.

	if value, ok := functions[cmd]; ok {
		value.(func(context.Context, []string))(ctx, args)
	} else {
		help.ShowHelp(usage)
	}
}