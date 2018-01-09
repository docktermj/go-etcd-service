package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/docktermj/go-logger/logger"
	"github.com/docktermj/go-etcd-service/common/runner"
	"github.com/docktermj/go-etcd-service/service/etcd"
	"github.com/docopt/docopt-go"
	"github.com/spf13/viper"
)

var (
	programName    = "go-etcd-service"
	buildVersion   = "0.0.0"
	buildIteration = "0"
	functions      = map[string]interface{}{}
	services       = []interface{}{
		etcd.ServiceWithWaitGroup,
	}
)

func defineConfiguration(args map[string]interface{}) {

	// List of 1 or more client endpoints for this instance of Etcd.

	key := "etcdClientEndpoints"
	viper.SetDefault(key, "http://localhost:2379")
	viper.BindEnv(key, "ETCD_CLIENT_ENDPOINTS")
	rawResult := args["--client-endpoints"]
	if rawResult != nil {
		viper.Set(key, rawResult.(string)) // Work-around for Bug #369
	}

	// List of 1 or more peer endpoints for this instance of Etcd.

	key = "etcdPeerEndpoints"
	viper.SetDefault(key, "http://localhost:2380")
	viper.BindEnv(key, "ETCD_PEER_ENDPOINTS")
	rawResult = args["--peer-endpoints"]
	if rawResult != nil {
		viper.Set(key, rawResult.(string)) // Work-around for Bug #369
	}

	// List of 0 or more client endpoints for other instances of Etcd in the cluster.

	key = "etcdClusterClientEndpoints"
	viper.SetDefault(key, "")
	viper.BindEnv(key, "ETCD_CLUSTER_CLIENT_ENDPOINTS")
	rawResult = args["--cluster-client-endpoints"]
	if rawResult != nil {
		viper.Set(key, rawResult.(string)) // Work-around for Bug #369
	}
}

func main() {

	usage := `
Usage:
    go-etcd-service [<command>] [options]

Options:
   -h, --help                         Show this help
   --client-endpoints=<list>          Port for clients to access. Default: http://localhost:2379
   --peer-endpoints=<list>            Port for peers to access. Default: http://localhost:2380
   --cluster-client-endpoints=<list>  Optional: List of ...
`

	// DocOpt processing.

	commandVersion := fmt.Sprintf("%s %s-%s", programName, buildVersion, buildIteration)
	args, _ := docopt.Parse(usage, nil, true, commandVersion, false)
	defineConfiguration(args)

	// Configure output log.

	log.SetFlags(log.Lshortfile | log.Ldate | log.Lmicroseconds | log.LUTC)
	logger.SetLevel(logger.LevelDebug)

	// Create top-level context.

	topContext, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl-c interrupts to cancel the context.

	interruptChannel := make(chan os.Signal, 2)
	signal.Notify(interruptChannel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-interruptChannel
		cancel()
	}()

	// Show debugging information.

	if logger.IsDebug() {
		logger.Debugf("os.Args: %+v\n", os.Args)
		logger.Debugf("args: %+v\n", args)
		logger.Debugf("topContext: %+v\n", topContext)
	}

	// If subcommand was specified, handle it and exit.

	if args["<command>"] != nil {
		_, hasSubcommand := functions[args["<command>"].(string)]
		if hasSubcommand {
			argv := os.Args[1:]
			runner.Run(topContext, argv, functions, usage)
			cancel()
			os.Exit(0)
		}
	}

	// Setup service synchronization.

	waitGroup := sync.WaitGroup{}

	// Start services.

	for _, service := range services {
		fn := service.(func(context.Context, *sync.WaitGroup) error)
		go func() {
			err := fn(topContext, &waitGroup)
			if err != nil {
				logger.Errorf("error starting service - %s", err.Error())
			}
		}()
	}

	// Wait until all processes are done or termination.

	<-topContext.Done()
	waitGroup.Wait()

	// Epilog.

	logger.Debugf("Done\n")
}
