package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

const (
	TailSleep         = 5
	ResponseTimeout   = 10
	AuthTypeAWS       = "aws"
	AuthTypeBasic     = "basic"
	SearchRequestSize = 10000
)

var (
	version = ""
	commit  = ""
	date    = ""
)

var DebugLogger = log.New(io.Discard, "DEBUG: ", log.Ldate|log.Ltime|log.Lshortfile)

func FatalError(err error) {
	fmt.Fprintf(os.Stderr, "clibana: %v\n", err)
	os.Exit(1)
}

func main() {
	clibanaConfig := NewClibanaConfig()

	if clibanaConfig.Debug {
		DebugLogger.SetOutput(os.Stderr)

	}

	client, err := createClient(clibanaConfig)
	if err != nil {
		FatalError(fmt.Errorf("Failed to create OpenSearch client: %w", err))
	}

	DebugLogger.Printf("Configuration: %+v\n", clibanaConfig)

	switch {
	case clibanaConfig.Search != nil:
		search(client, clibanaConfig)
	case clibanaConfig.Mappings != nil:
		mappings(client, clibanaConfig)
	case clibanaConfig.Indices != nil:
		indices(client, clibanaConfig)
	default:
		FatalError(fmt.Errorf("no subcommand specified"))
	}

}
