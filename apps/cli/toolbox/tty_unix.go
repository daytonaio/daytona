//go:build unix

package toolbox

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/gorilla/websocket"
	"golang.org/x/term"
)

func (c *Client) setupSignalHandling(sigChan chan os.Signal, ws *websocket.Conn) {
	// Handle SIGWINCH (terminal resize), SIGINT, and SIGTERM on Unix systems
	signal.Notify(sigChan, syscall.SIGWINCH, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for sig := range sigChan {
			switch sig {
			case syscall.SIGWINCH:
				// Handle terminal resize
				c.handleResize(ws)
			case syscall.SIGINT, syscall.SIGTERM:
				// Forward interrupt signal to the remote process
				// Send Ctrl+C (0x03) to the WebSocket
				if sig == syscall.SIGINT {
					ws.WriteMessage(websocket.BinaryMessage, []byte{0x03})
				}
			}
		}
	}()
}

func (c *Client) handleResize(ws *websocket.Conn) {
	// Get current terminal size
	cols, rows, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		return // Skip resize on error
	}

	// Send resize message to the WebSocket
	resizeMsg := map[string]interface{}{
		"type": "resize",
		"cols": cols,
		"rows": rows,
	}

	// Best effort - don't block on resize
	ws.WriteJSON(resizeMsg)
}
