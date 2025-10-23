package main

import (
	"io"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/daytonaio/mcp/internal/config"
	"github.com/daytonaio/mcp/internal/server"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"

	log "github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Errorf("Failed to get config: %v", err)
		return
	}

	server := server.NewMCPServer(server.MCPServerConfig{
		Port:            cfg.Port,
		TLSCertFilePath: cfg.TLSCertFilePath,
		TLSKeyFilePath:  cfg.TLSKeyFilePath,
		ApiUrl:          cfg.ApiUrl,
	})

	mcpServerErrChan := make(chan error)

	go func() {
		mcpServerErrChan <- server.Start()
	}()

	interruptChan := make(chan os.Signal, 1)
	signal.Notify(interruptChan, os.Interrupt)

	select {
	case err := <-mcpServerErrChan:
		log.Errorf("MCP server error: %v", err)
		return
	case <-interruptChan:
		server.Stop()
	}
}

func init() {
	// Load .env file
	err := godotenv.Load()
	if err != nil {
		log.Printf("Error loading .env file: %v", err)
		// Continue anyway, as environment variables might be set directly
	}

	logLevel := log.WarnLevel

	logLevelEnv, logLevelSet := os.LookupEnv("LOG_LEVEL")

	if logLevelSet {
		var err error
		logLevel, err = log.ParseLevel(logLevelEnv)
		if err != nil {
			log.Warnf("Failed to parse log level '%s', using WarnLevel: %v", logLevelEnv, err)
			logLevel = log.WarnLevel
		}
	}

	log.SetLevel(logLevel)
	log.SetOutput(os.Stdout)

	logFilePath, logFilePathSet := os.LookupEnv("LOG_FILE_PATH")
	if logFilePathSet {
		logDir := filepath.Dir(logFilePath)

		if err := os.MkdirAll(logDir, 0755); err != nil {
			log.Errorf("Failed to create log directory: %v", err)
			os.Exit(1)
		}

		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Errorf("Failed to open log file: %v", err)
			os.Exit(1)
		}

		log.SetOutput(io.MultiWriter(os.Stdout, file))
	}

	zerologLevel, err := zerolog.ParseLevel(logLevel.String())
	if err != nil {
		log.Warnf("Failed to parse zerolog level, using ErrorLevel: %v", err)
		zerologLevel = zerolog.ErrorLevel
	}

	zerolog.SetGlobalLevel(zerologLevel)
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
}
