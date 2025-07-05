// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package terminal

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"

	"github.com/daytonaio/daemon/pkg/common"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Be careful with this in production
	},
}

type windowSize struct {
	Rows uint16 `json:"rows"`
	Cols uint16 `json:"cols"`
}

func StartTerminalServer(port int) error {
	// Prepare the embedded frontend files
	// Serve the files from the embedded filesystem
	staticFS, err := fs.Sub(static, "static")
	if err != nil {
		return err
	}

	http.Handle("/", http.FileServer(http.FS(staticFS)))
	http.HandleFunc("/ws", handleWebSocket)

	addr := fmt.Sprintf(":%d", port)
	log.Printf("Starting terminal server on http://localhost%s", addr)
	return http.ListenAndServe(addr, nil)
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	// Create a new UTF8Decoder instance for this connection
	decoder := NewUTF8Decoder()

	sizeCh := make(chan common.TTYSize)
	stdInReader, stdInWriter := io.Pipe()
	stdOutReader, stdOutWriter := io.Pipe()

	// Handle websocket -> pty
	go func() {
		for {
			messageType, p, err := conn.ReadMessage()
			if err != nil {
				return
			}

			// Check if it's a resize message
			if messageType == websocket.TextMessage {
				var size windowSize
				if err := json.Unmarshal(p, &size); err == nil {
					sizeCh <- common.TTYSize{
						Height: int(size.Rows),
						Width:  int(size.Cols),
					}
					continue
				}
			}

			// Write to pty
			_, err = stdInWriter.Write(p)
			if err != nil {
				return
			}
		}
	}()

	go func() {
		// Handle pty -> websocket
		buf := make([]byte, 1024)
		for {
			n, err := stdOutReader.Read(buf)
			if err != nil {
				if err != io.EOF {
					log.Printf("Failed to read from pty: %v", err)
				}
				return
			}

			// A multi-byte UTF-8 character can be split across stream reads.
			// UTF8Decoder buffers incomplete sequences to ensure proper decoding.
			decoded := decoder.Write(buf[:n])

			err = conn.WriteMessage(websocket.TextMessage, []byte(decoded))
			if err != nil {
				log.Printf("Failed to write to websocket: %v", err)
				return
			}
		}
	}()

	// Create a pty
	err = common.SpawnTTY(common.SpawnTTYOptions{
		Dir:    "/",
		StdIn:  stdInReader,
		StdOut: stdOutWriter,
		Term:   "xterm-256color",
		SizeCh: sizeCh,
	})
	if err != nil {
		log.Printf("Failed to start pty: %v", err)
		return
	}
}
