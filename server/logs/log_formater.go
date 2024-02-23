package logs

import (
	"os"
	"path"

	"github.com/daytonaio/daytona/server/config"
	log "github.com/sirupsen/logrus"
)

type LogFormatter struct {
	textFormater *log.TextFormatter
	LogFilePath  *string
}

func (f *LogFormatter) Format(entry *log.Entry) ([]byte, error) {
	formatted, err := f.textFormater.Format(entry)
	if err != nil {
		return nil, err
	}

	if f.LogFilePath != nil {
		// Write to file
		file, err := os.OpenFile(*f.LogFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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

	if _, err := os.Stat(filePath); err == nil {
		os.Remove(filePath)
	}

	logFormatter := &LogFormatter{
		textFormater: new(log.TextFormatter),
		LogFilePath:  &filePath,
	}

	log.SetFormatter(logFormatter)

	return nil
}
