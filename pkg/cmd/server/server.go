package server

import (
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/daytonaio/daytona/pkg/cmd/server/daemon"
	"github.com/daytonaio/daytona/pkg/server"
	"github.com/daytonaio/daytona/pkg/server/config"
	"github.com/daytonaio/daytona/pkg/server/frpc"
	"github.com/daytonaio/daytona/pkg/views/util"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var runAsDaemon bool

var ServerCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the server process",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if log.GetLevel() < log.InfoLevel {
			log.SetLevel(log.InfoLevel)
		}

		if runAsDaemon {
			fmt.Println("Starting the Daytona Server daemon...")
			daemon.Start()
			return
		}

		router := mux.NewRouter()
		router.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
			if err := performHealthCheck(); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Server unhealthy: %v", err)
				return
			}
			w.WriteHeader(http.StatusOK)
			fmt.Fprintf(w, "Server healthy")
		})

		errCh := make(chan error)
		err := server.Start(errCh)
		if err != nil {
			log.Fatal(err)
		}

		c, err := config.GetConfig()
		if err != nil {
			log.Fatal(err)
		}

		go func() {
			if err := performHealthCheck(); err != nil {
				log.Error(err)
				errCh <- err
			}
		}()

		select {
		case err := <-errCh:
			log.Fatal(err)
		case <-time.After(5 * time.Second):
			util.RenderBorderedMessage(fmt.Sprintf("Daytona Server running on port: %d.\nTo connect to the server remotely, use the following command on the client machine:\n\ndaytona profile add -a %s", c.ApiPort, frpc.GetApiUrl(c)))
		}

		err = <-errCh
		if err != nil {
			log.Fatal(err)
		}
	},
}

func performHealthCheck() error {
	c, err := config.GetConfig()
	if err != nil {
		return err
	}
	addr := fmt.Sprintf(":%d", c.ApiPort)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("server not listening on port %d: %v", c.ApiPort, err)
	}
	conn.Close()
	return nil
}

func init() {
	ServerCmd.PersistentFlags().BoolVarP(&runAsDaemon, "daemon", "d", false, "Run the server as a daemon")
	ServerCmd.AddCommand(configureCmd)
	ServerCmd.AddCommand(configCmd)
	ServerCmd.AddCommand(logsCmd)
	ServerCmd.AddCommand(stopCmd)
	ServerCmd.AddCommand(restartCmd)
}
