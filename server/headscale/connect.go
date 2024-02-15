package headscale

import (
	"fmt"
	"html"
	"net/http"

	"tailscale.com/tsnet"

	"github.com/daytonaio/daytona/server/config"
	"github.com/daytonaio/daytona/server/frpc"
	log "github.com/sirupsen/logrus"
)

var s = &tsnet.Server{
	Hostname: "server",
}

func Connect() error {
	c, err := config.GetConfig()
	if err != nil {
		return err
	}

	err = CreateUser()
	if err != nil {
		log.Fatal(err)
	}

	authKey, err := CreateAuthKey()
	if err != nil {
		log.Fatal(err)
	}

	s.ControlURL = frpc.GetServerUrl(c)
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

	return http.Serve(ln, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
	}))
}

func HTTPClient() *http.Client {
	return s.HTTPClient()
}
