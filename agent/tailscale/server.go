package tailscale

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/netip"
	"os"

	"github.com/daytonaio/daytona/agent/config"
	"tailscale.com/tsnet"
)

func Start(c *config.Config) {
	flag.Parse()
	s := new(tsnet.Server)
	s.Hostname = fmt.Sprintf("%s-%s", os.Getenv("DAYTONA_WS_ID"), os.Getenv("DAYTONA_WS_PROJECT_NAME"))
	s.ControlURL = c.Server.Url
	s.AuthKey = c.Server.AuthKey

	defer s.Close()
	ln, err := s.Listen("tcp", ":80")
	if err != nil {
		log.Fatal(err)
	}

	defer ln.Close()

	s.RegisterFallbackTCPHandler(func(src, dest netip.AddrPort) (handler func(net.Conn), intercept bool) {
		destPort := dest.Port()

		return func(src net.Conn) {
			defer src.Close()
			dst, err := net.Dial("tcp", fmt.Sprintf("localhost:%d", destPort))
			if err != nil {
				log.Printf("Dial failed: %v", err)
				return
			}
			defer dst.Close()

			done := make(chan struct{})

			go func() {
				defer src.Close()
				defer dst.Close()
				io.Copy(dst, src)
				done <- struct{}{}
			}()

			go func() {
				defer src.Close()
				defer dst.Close()
				io.Copy(src, dst)
				done <- struct{}{}
			}()

			<-done
			<-done
		}, true
	})

	log.Fatal(http.Serve(ln, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Ok\n")
	})))
}
