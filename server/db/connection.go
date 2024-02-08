// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package db

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/glebarez/sqlite"
	log "github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var _db *gorm.DB

func getConnection() *gorm.DB {
	if _db != nil {
		return _db
	}

	dbPath, err := getDbPath()
	if err != nil {
		panic(err)
	}

	log.Debug(fmt.Sprintf("Opening db at %s", dbPath))

	logLevel := log.ErrorLevel
	if log.GetLevel() == log.DebugLevel {
		logLevel = log.DebugLevel
	}

	newLogger := logger.New(
		log.New(),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.LogLevel(logLevel),
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  true,
		},
	)

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		panic("failed to connect database")
	}

	_db = db

	return _db
}

func getDbPath() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	dir := path.Join(userConfigDir, "daytona")

	err = os.MkdirAll(dir, 0755)
	if err != nil {
		return "", err
	}

	return path.Join(dir, "db"), nil
}
