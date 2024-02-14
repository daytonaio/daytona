// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package server

import (
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"time"

	proto "github.com/daytonaio/daytona/common/grpc/proto"
	"github.com/daytonaio/daytona/server/config"
	plugin_grpc "github.com/daytonaio/daytona/server/grpc/plugins"
	ports_grpc "github.com/daytonaio/daytona/server/grpc/ports"
	server_grpc "github.com/daytonaio/daytona/server/grpc/server"
	workspace_grpc "github.com/daytonaio/daytona/server/grpc/workspace"
	"github.com/daytonaio/daytona/server/headscale"
	"github.com/daytonaio/daytona/server/ssh_gateway"
	"github.com/hashicorp/go-plugin"
	"tailscale.com/tsnet"

	"google.golang.org/grpc"

	log "github.com/sirupsen/logrus"
)

type Self struct {
	HostName string `json:"HostName"`
	DNSName  string `json:"DNSName"`
}

func Start() error {
	log.Info("Starting Daytona server")

	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	_, err = config.GetWorkspaceKey()
	if os.IsNotExist(err) {
		log.Info("Generating workspace key")
		err = config.GenerateWorkspaceKey()
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	var lis *net.Listener

	lis, err = getTcpListener()
	if err != nil {
		return err
	}
	defer (*lis).Close()

	s := grpc.NewServer()
	workspaceServer := &workspace_grpc.WorkspaceServer{}
	proto.RegisterWorkspaceServiceServer(s, workspaceServer)
	portsServer := &ports_grpc.PortsServer{}
	proto.RegisterPortsServer(s, portsServer)
	serverGrpcServer := &server_grpc.ServerGRPCServer{}
	proto.RegisterServerServer(s, serverGrpcServer)
	pluginsServer := &plugin_grpc.PluginsServer{}
	proto.RegisterPluginsServer(s, pluginsServer)

	err = downloadDefaultPlugins()
	if err != nil {
		return err
	}

	err = registerProvisioners(c)
	if err != nil {
		return err
	}
	err = registerAgentServices(c)
	if err != nil {
		return err
	}

	log.Infof("Daytona server started %v", (*lis).Addr())

	go func() {
		if err := ssh_gateway.Start(); err != nil {
			log.Error(err)
		}
	}()

	go func() {
		interruptChannel := make(chan os.Signal, 1)
		signal.Notify(interruptChannel, os.Interrupt)

		for range interruptChannel {
			log.Info("Shutting down")
			plugin.CleanupClients()
			os.Exit(0)
		}
	}()

	go func() {
		time.Sleep(5 * time.Second)
		err := headscale.CreateUser()
		if err != nil {
			log.Fatal(err)
		}

		authKey, err := headscale.CreateAuthKey()
		if err != nil {
			log.Fatal(err)
		}

		s := new(tsnet.Server)
		s.Hostname = "server"
		s.ControlURL = "https://toma.frps.daytona.io"
		s.AuthKey = authKey

		defer s.Close()
		ln, err := s.Listen("tcp", ":80")
		if err != nil {
			log.Fatal(err)
		}

		defer ln.Close()

		lc, err := s.LocalClient()
		if err != nil {
			log.Fatal(err)
		}

		log.Fatal(http.Serve(ln, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			who, err := lc.WhoIs(r.Context(), r.RemoteAddr)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			fmt.Fprintf(w, "<html><body><h1>Hello, tailnet!</h1>\n")
			fmt.Fprintf(w, "<p>You are <b>%s</b> from <b>%s</b> (%s)</p>",
				html.EscapeString(who.UserProfile.LoginName),
				html.EscapeString(who.Node.ComputedName),
				r.RemoteAddr)
		})))
	}()

	go func() {
		for {
			time.Sleep(5 * time.Second)
			req, err := http.Get("http://wrk1-tpuljak:3000")
			if err != nil {
				log.Error(err)
				continue
			}
			body, err := io.ReadAll(req.Body)
			if err != nil {
				log.Error(err)
				continue
			}
			log.Info(string(body))
			req.Body.Close()
		}
	}()

	go func() {
		log.Fatal(headscale.Start())
	}()

	if err := s.Serve(*lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	return nil
}

func StartDaemon() error {

	scriptFile, err := createTemporaryScriptFile()
	if err != nil {
		log.Error(err)
		return nil
	}
	defer func() {
		scriptFile.Close()
		os.Remove(scriptFile.Name())
	}()

	scriptPath := scriptFile.Name()

	// Run the bash script and capture its output
	cmd := exec.Command("bash", scriptPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Error(string(output))
		return err
	}
	log.Info(string(output))

	return nil
}

func getTcpListener() (*net.Listener, error) {
	listener, err := net.Listen("tcp", "0.0.0.0:2790")
	if err != nil {
		return nil, err
	}
	return &listener, nil
}
