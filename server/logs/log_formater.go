package logs

import (
	"os"
	"path"

	"github.com/daytonaio/daytona/server/config"
	log "github.com/sirupsen/logrus"
)

var LogFilePath *string

type logFormatter struct {
	textFormater *log.TextFormatter
}

func (f *logFormatter) Format(entry *log.Entry) ([]byte, error) {
	formatted, err := f.textFormater.Format(entry)
	if err != nil {
		return nil, err
	}

	if LogFilePath != nil {
		// Write to file
		file, err := os.OpenFile(*LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		_, err = file.Write(formatted)
		if err != nil {
			return nil, err
		}
	}

	return formatted, nil
}

func Init() error {
	configDir, err := config.GetConfigDir()
	if err != nil {
		return err
	}

	filePath := path.Join(configDir, "daytona.log")
	LogFilePath = &filePath

	if _, err := os.Stat(*LogFilePath); err == nil {
		os.Remove(*LogFilePath)
	}

	logFormatter := &logFormatter{
		textFormater: new(log.TextFormatter),
	}

	log.SetFormatter(logFormatter)

	return nil
}
