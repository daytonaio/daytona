package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	golog "log"

	"github.com/daytonaio/daytona/cmd/daytona/config"
	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/cmd"
	"github.com/daytonaio/daytona/pkg/cmd/workspacemode"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Define a new flag for editing SSH config
	edit := flag.Bool("edit", false, "Edit SSH config for a specific project")
	projectHostname := flag.String("project", "", "Specify the project hostname")
	flag.Parse()

	// Check if the edit flag is set
	if *edit {
		if *projectHostname == "" {
			log.Fatal("Error: Project hostname must be specified with -edit option")
		}
		err := config.EditSSHConfig(*projectHostname)
		if err != nil {
			log.Fatalf("Error editing SSH config: %v", err)
		}
		return
	}

	if internal.WorkspaceMode() {
		err := workspacemode.Execute()
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}

	// Example: Update usage instructions if present
	fmt.Println("Usage: dtn ssh [options] <project>")
}

func init() {
	logLevel := log.WarnLevel

	logLevelEnv, logLevelSet := os.LookupEnv("LOG_LEVEL")

	if logLevelSet {
		var err error
		logLevel, err = log.ParseLevel(logLevelEnv)
		if err != nil {
			logLevel = log.WarnLevel
		}
	}

	log.SetLevel(logLevel)

	zerologLevel, err := zerolog.ParseLevel(logLevel.String())
	if err != nil {
		zerologLevel = zerolog.ErrorLevel
	}

	zerolog.SetGlobalLevel(zerologLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zlog.Logger = zlog.Output(zerolog.ConsoleWriter{
		Out:        &util.DebugLogWriter{},
		TimeFormat: time.RFC3339,
	})

	golog.SetOutput(&util.DebugLogWriter{})
}
