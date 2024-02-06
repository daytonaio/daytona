// Copyright 2024 Daytona Platforms Inc.
// SPDX-License-Identifier: Apache-2.0

package cmd

import (
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Exposes a port in project container
var exposePortCmd = &cobra.Command{
	Use:    "expose-port [PORT]",
	Short:  "Expose port in project container",
	Args:   cobra.ExactArgs(1),
	Hidden: true,
	Run: func(cmd *cobra.Command, args []string) {
		portStr := args[0]

		port, err := strconv.Atoi(portStr)
		if err != nil {
			log.Fatal("Error: " + portStr + " is not a valid port number")
		}

		// Handle Ctrl+C to gracefully close the connection
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

		conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			log.Printf("Failed to connect: %v\n", err)
			return
		}
		defer conn.Close()

		go func() {
			io.Copy(conn, os.Stdin)
		}()

		go func() {
			io.Copy(os.Stdout, conn)
		}()

		<-sigCh
	},
}
