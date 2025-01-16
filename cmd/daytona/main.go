// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"
	"time"

	golog "log"

	"github.com/daytonaio/daytona/internal"
	"github.com/daytonaio/daytona/internal/util"
	"github.com/daytonaio/daytona/pkg/cmd"
	"github.com/daytonaio/daytona/pkg/cmd/workspacemode"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	log "github.com/sirupsen/logrus"
)

func main() {
	// err := Session13()
	// if err != nil {
	// 	log.Fatal(err)
	// } else {
	// 	return
	// }

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

// func Session13() error {
// 	sessId := "123"
// 	CreateSession(process.CreateSessionRequest{SessionId: sessId})

// 	cmd := "ls -la"
// 	output, err := SessionExec(sessId, cmd)
// 	if err != nil {
// 		return err
// 	}
// 	log.Infof("Output: %s", output)

// 	cmd2 := "cd /home"
// 	output2, err := SessionExec(sessId, cmd2)
// 	if err != nil {
// 		return err
// 	}
// 	log.Infof("Output2: %s", output2)

// 	cmd3 := "ls -la"
// 	output3, err := SessionExec(sessId, cmd3)
// 	if err != nil {
// 		return err
// 	}
// 	log.Infof("Output3: %s", output3)

// 	err = DeleteSession(sessId)
// 	if err != nil {
// 		return err
// 	}

// 	sessId2 := "456"
// 	CreateSession(process.CreateSessionRequest{SessionId: sessId2})

// 	cmd4 := "echo $VAR"
// 	output4, err := SessionExec(sessId2, cmd4)
// 	if err != nil {
// 		return err
// 	}
// 	log.Infof("Output4: %s", output4)

// 	return nil
// }

// type Session struct {
// 	Cmd          *exec.Cmd
// 	StdoutReader *bufio.Reader
// 	StdinWriter  io.Writer
// }

// var sessions = map[string]*Session{}

// func CreateSession(req process.CreateSessionRequest) {
// 	cmd := exec.Command("/bin/sh")

// 	stdinWriter, err := cmd.StdinPipe()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	stdoutReader, err := cmd.StdoutPipe()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	err = cmd.Start()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	configDir, err := config.GetConfigDir()
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	err = os.MkdirAll(filepath.Join(configDir, "sessions"), 0755)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	sessions[req.SessionId] = &Session{
// 		Cmd:          cmd,
// 		StdoutReader: bufio.NewReader(stdoutReader),
// 		StdinWriter:  stdinWriter,
// 	}
// }

// func SessionExec(sessionId, cmd string) (string, error) {
// 	configDir, err := config.GetConfigDir()
// 	if err != nil {
// 		return "", err
// 	}

// 	cmdId := uuid.NewString()
// 	fmt.Println("CMD: ", cmd, " cmdId: ", cmdId)

// 	err = os.MkdirAll(filepath.Join(configDir, "sessions", sessionId, cmdId), 0755)
// 	if err != nil {
// 		return "", err
// 	}

// 	logFile, err := os.Create(filepath.Join(configDir, "sessions", sessionId, cmdId, "output.log"))
// 	if err != nil {
// 		return "", err
// 	}
// 	defer logFile.Close()

// 	session, ok := sessions[sessionId]
// 	if !ok {
// 		return "", errors.New("session not found")
// 	}

// 	cmdToExec := fmt.Sprintf("%s && echo DAYTONA_CMD_END\n\n", cmd)

// 	output := make(chan string)
// 	outputErr := make(chan error)
// 	defer close(output)
// 	defer close(outputErr)

// 	go func() {
// 		out := ""
// 		for {
// 			line, _, err := session.StdoutReader.ReadLine()
// 			if err != nil {
// 				if err == io.EOF {
// 					break
// 				}
// 				outputErr <- err
// 			}
// 			if strings.Contains(string(line), "DAYTONA_CMD_END") {
// 				break
// 			}

// 			l := string(line) + "\n"
// 			out += l
// 			logFile.Write([]byte(l))
// 		}

// 		output <- out
// 		outputErr <- nil
// 	}()

// 	_, err = session.StdinWriter.Write([]byte(cmdToExec))
// 	if err != nil {
// 		return "", err
// 	}

// 	return <-output, <-outputErr
// }

// func DeleteSession(sessionId string) error {
// 	session, ok := sessions[sessionId]
// 	if !ok {
// 		return errors.New("session not found")
// 	}

// 	err := session.Cmd.Process.Kill()
// 	if err != nil {
// 		return err
// 	}

// 	delete(sessions, sessionId)

// 	configDir, err := config.GetConfigDir()
// 	if err != nil {
// 		return err
// 	}

// 	return os.RemoveAll(filepath.Join(configDir, "sessions", sessionId))
// }
