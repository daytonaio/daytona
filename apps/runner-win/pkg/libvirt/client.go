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

	// Initialize required directories on the libvirt host
	if err := l.initDirectories(); err != nil {
		log.Warnf("Failed to initialize directories: %v", err)
		// Don't fail startup, directories might already exist or be created manually
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

	// Check if connection exists and is still alive
	needsReconnect := conn == nil
	if conn != nil {
		alive, err := conn.IsAlive()
		if err != nil || !alive {
			log.Warnf("Libvirt connection is dead, will reconnect: %v", err)
			needsReconnect = true
			// Close the dead connection
			l.connMutex.Lock()
			if l.conn != nil {
				l.conn.Close()
				l.conn = nil
			}
			l.connMutex.Unlock()
		}
	}

	if needsReconnect {
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

// GetURI returns the libvirt connection URI
func (l *LibVirt) GetURI() string {
	return l.libvirtURI
}

// GetSSHHost extracts user@host from the libvirt URI for SSH connections
// Returns empty string if the URI is not an SSH connection
func (l *LibVirt) GetSSHHost() string {
	uri := l.libvirtURI
	if len(uri) == 0 {
		return ""
	}

	// Find the start after "://"
	startIdx := -1
	for i := 0; i < len(uri)-2; i++ {
		if uri[i] == ':' && uri[i+1] == '/' && uri[i+2] == '/' {
			startIdx = i + 3
			break
		}
	}
	if startIdx == -1 {
		return ""
	}

	// Find the end before the path "/"
	endIdx := len(uri)
	for i := startIdx; i < len(uri); i++ {
		if uri[i] == '/' {
			endIdx = i
			break
		}
	}

	return uri[startIdx:endIdx]
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

// LookupDomainBySandboxId looks up a domain by sandbox ID.
// It first tries to find the domain by UUID, then falls back to looking up by name.
func (l *LibVirt) LookupDomainBySandboxId(conn *libvirt.Connect, sandboxId string) (*libvirt.Domain, error) {
	// First try to look up by UUID string
	domain, err := conn.LookupDomainByUUIDString(sandboxId)
	if err != nil {
		// Fallback to looking up by name
		domain, err = conn.LookupDomainByName(sandboxId)
		if err != nil {
			return nil, fmt.Errorf("domain not found by UUID or name: %s: %w", sandboxId, err)
		}
	}
	return domain, nil
}
