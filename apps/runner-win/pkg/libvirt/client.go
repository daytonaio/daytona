// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package libvirt

import (
	"fmt"
	"io"
	"sync"

	"github.com/daytonaio/runner-win/pkg/cache"
	"github.com/daytonaio/runner-win/pkg/netrules"
	log "github.com/sirupsen/logrus"
	"libvirt.org/go/libvirt"
)

type LibVirtConfig struct {
	LibvirtURI             string // e.g., "qemu+ssh://root@h1001.blinkbox.dev/system"
	StatesCache            *cache.StatesCache
	LogWriter              io.Writer
	AWSRegion              string
	AWSEndpointUrl         string
	AWSAccessKeyId         string
	AWSSecretAccessKey     string
	DaemonPath             string
	ComputerUsePluginPath  string
	NetRulesManager        *netrules.NetRulesManager
	ResourceLimitsDisabled bool
	DaemonStartTimeoutSec  int
	SandboxStartTimeoutSec int
	UseSnapshotEntrypoint  bool
}

func NewLibVirt(config LibVirtConfig) (*LibVirt, error) {
	if config.DaemonStartTimeoutSec <= 0 {
		log.Warnf("Invalid DaemonStartTimeoutSec value: %d. Using default value: 60 seconds", config.DaemonStartTimeoutSec)
		config.DaemonStartTimeoutSec = 60
	}

	if config.SandboxStartTimeoutSec <= 0 {
		log.Warnf("Invalid SandboxStartTimeoutSec value: %d. Using default value: 30 seconds", config.SandboxStartTimeoutSec)
		config.SandboxStartTimeoutSec = 30
	}

	if config.LibvirtURI == "" {
		return nil, fmt.Errorf("LibvirtURI is required")
	}

	l := &LibVirt{
		libvirtURI:             config.LibvirtURI,
		statesCache:            config.StatesCache,
		logWriter:              config.LogWriter,
		awsRegion:              config.AWSRegion,
		awsEndpointUrl:         config.AWSEndpointUrl,
		awsAccessKeyId:         config.AWSAccessKeyId,
		awsSecretAccessKey:     config.AWSSecretAccessKey,
		daemonPath:             config.DaemonPath,
		computerUsePluginPath:  config.ComputerUsePluginPath,
		netRulesManager:        config.NetRulesManager,
		resourceLimitsDisabled: config.ResourceLimitsDisabled,
		daemonStartTimeoutSec:  config.DaemonStartTimeoutSec,
		sandboxStartTimeoutSec: config.SandboxStartTimeoutSec,
		useSnapshotEntrypoint:  config.UseSnapshotEntrypoint,
		domainMutexes:          make(map[string]*sync.Mutex),
	}

	// Establish connection
	if err := l.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}

	return l, nil
}

func (l *LibVirt) connect() error {
	l.connMutex.Lock()
	defer l.connMutex.Unlock()

	if l.conn != nil {
		// Already connected
		return nil
	}

	log.Infof("Connecting to libvirt URI: %s", l.libvirtURI)
	conn, err := libvirt.NewConnect(l.libvirtURI)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", l.libvirtURI, err)
	}

	l.conn = conn
	log.Infof("Successfully connected to libvirt")
	return nil
}

func (l *LibVirt) Close() error {
	l.connMutex.Lock()
	defer l.connMutex.Unlock()

	if l.conn != nil {
		if _, err := l.conn.Close(); err != nil {
			return fmt.Errorf("failed to close libvirt connection: %w", err)
		}
		l.conn = nil
	}
	return nil
}

func (l *LibVirt) getConnection() (*libvirt.Connect, error) {
	l.connMutex.RLock()
	conn := l.conn
	l.connMutex.RUnlock()

	if conn == nil {
		// Try to reconnect
		if err := l.connect(); err != nil {
			return nil, err
		}
		l.connMutex.RLock()
		conn = l.conn
		l.connMutex.RUnlock()
	}

	return conn, nil
}

func (l *LibVirt) ApiClient() *libvirt.Connect {
	conn, err := l.getConnection()
	if err != nil {
		log.Errorf("Failed to get libvirt connection: %v", err)
		return nil
	}
	return conn
}

func (l *LibVirt) getDomainMutex(domainName string) *sync.Mutex {
	l.domainMutexesMutex.Lock()
	defer l.domainMutexesMutex.Unlock()

	if _, ok := l.domainMutexes[domainName]; !ok {
		l.domainMutexes[domainName] = &sync.Mutex{}
	}

	return l.domainMutexes[domainName]
}

type LibVirt struct {
	conn                   *libvirt.Connect
	connMutex              sync.RWMutex
	libvirtURI             string
	statesCache            *cache.StatesCache
	logWriter              io.Writer
	awsRegion              string
	awsEndpointUrl         string
	awsAccessKeyId         string
	awsSecretAccessKey     string
	daemonPath             string
	computerUsePluginPath  string
	netRulesManager        *netrules.NetRulesManager
	resourceLimitsDisabled bool
	daemonStartTimeoutSec  int
	sandboxStartTimeoutSec int
	useSnapshotEntrypoint  bool
	domainMutexes          map[string]*sync.Mutex
	domainMutexesMutex     sync.Mutex
}
