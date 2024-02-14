package tailscale

import (
	"flag"
	"fmt"
	"html"
	"io"
	"log"
	"net"
	"net/http"
	"net/netip"
	"os"
	"strings"

	"tailscale.com/tsnet"
)

func Start(authKey string) {
	flag.Parse()
	s := new(tsnet.Server)
	s.Hostname = fmt.Sprintf("%s-%s", os.Getenv("DAYTONA_WS_ID"), os.Getenv("DAYTONA_WS_PROJECT_NAME"))
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

	s.RegisterFallbackTCPHandler(func(src, dest netip.AddrPort) (handler func(net.Conn), intercept bool) {
		log.Printf("fallback: %v -> %v", src, dest)
		// Get port from dst and return net conn from that port
		destPort := dest.Port()

		log.Printf("destPort: %d", destPort)

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
		who, err := lc.WhoIs(r.Context(), r.RemoteAddr)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}
		fmt.Fprintf(w, "<html><body><h1>Hello, tailnet!</h1>\n")
		fmt.Fprintf(w, "<p>You are <b>%s</b> from <b>%s</b> (%s)</p>",
			html.EscapeString(who.UserProfile.LoginName),
			html.EscapeString(firstLabel(who.Node.ComputedName)),
			r.RemoteAddr)
	})))
}

func firstLabel(s string) string {
	s, _, _ = strings.Cut(s, ".")
	return s
}
