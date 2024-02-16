package ssh

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"unsafe"

	"github.com/creack/pty"
	"github.com/gliderlabs/ssh"

	log "github.com/sirupsen/logrus"
)

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

func Start() {
	ForwardedTCPHandler := &ssh.ForwardedTCPHandler{}

	sshServer := ssh.Server{
		Addr: ":2222",
		Handler: func(s ssh.Session) {
			ptyReq, winCh, isPty := s.Pty()
			if isPty {
				cmd := exec.Command(getShell())
				cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
				f, err := pty.Start(cmd)
				if err != nil {
					panic(err)
				}
				go func() {
					for win := range winCh {
						setWinsize(f, win.Width, win.Height)
					}
				}()
				go func() {
					io.Copy(f, s) // stdin
				}()
				io.Copy(s, f) // stdout
				cmd.Wait()
			} else {
				defer s.Close()

				sshCommand := s.Command()

				if len(sshCommand) == 0 {
					return
				}

				cmd := exec.Command(sshCommand[0], sshCommand[1:]...)
				cmd.Stdin = s
				cmd.Stdout = s
				cmd.Stderr = s

				cmd.Run()
			}
		},
		ChannelHandlers: map[string]ssh.ChannelHandler{
			"session":      ssh.DefaultSessionHandler,
			"direct-tcpip": ssh.DirectTCPIPHandler, // ssh -L
		},
		RequestHandlers: map[string]ssh.RequestHandler{
			"tcpip-forward":        ForwardedTCPHandler.HandleSSHRequest,
			"cancel-tcpip-forward": ForwardedTCPHandler.HandleSSHRequest,
		},
		LocalPortForwardingCallback: ssh.LocalPortForwardingCallback(func(ctx ssh.Context, dhost string, dport uint32) bool {
			return true
		}),
		ReversePortForwardingCallback: ssh.ReversePortForwardingCallback(func(ctx ssh.Context, host string, port uint32) bool {
			return true
		}),
		SessionRequestCallback: func(sess ssh.Session, requestType string) bool {
			return true
		},
	}

	log.Info("starting ssh server on port 2222...")
	log.Fatal(sshServer.ListenAndServe())
}

func getShell() string {
	shellEnv, shellSet := os.LookupEnv("SHELL")

	if shellSet {
		return shellEnv
	}

	return "sh"
}
