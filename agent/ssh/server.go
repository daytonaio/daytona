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
	"github.com/pkg/sftp"
	crypto_ssh "golang.org/x/crypto/ssh"
	"golang.org/x/sys/unix"

	log "github.com/sirupsen/logrus"
)

func setWinsize(f *os.File, w, h int) {
	syscall.Syscall(syscall.SYS_IOCTL, f.Fd(), uintptr(syscall.TIOCSWINSZ),
		uintptr(unsafe.Pointer(&struct{ h, w, x, y uint16 }{uint16(h), uint16(w), 0, 0})))
}

func Start() {
	forwardedTCPHandler := &ssh.ForwardedTCPHandler{}

	sshServer := ssh.Server{
		Addr: ":2222",
		Handler: func(s ssh.Session) {
			ptyReq, winCh, isPty := s.Pty()
			if isPty {
				handlePty(s, ptyReq, winCh)
			} else {
				handleNonPty(s)
			}
		},
		ChannelHandlers: map[string]ssh.ChannelHandler{
			"session": ssh.DefaultSessionHandler,
			"direct-tcpip": func(srv *ssh.Server, conn *crypto_ssh.ServerConn, newChan crypto_ssh.NewChannel, ctx ssh.Context) {
				ssh.DirectTCPIPHandler(srv, conn, newChan, ctx)
			},
		},
		RequestHandlers: map[string]ssh.RequestHandler{
			"tcpip-forward":        forwardedTCPHandler.HandleSSHRequest,
			"cancel-tcpip-forward": forwardedTCPHandler.HandleSSHRequest,
		},
		SubsystemHandlers: map[string]ssh.SubsystemHandler{
			"sftp": sftpHandler,
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

	log.Println("starting ssh server on port 2222...")
	log.Fatal(sshServer.ListenAndServe())
}

func handlePty(s ssh.Session, ptyReq ssh.Pty, winCh <-chan ssh.Window) {
	cmd := exec.Command(getShell())

	if ssh.AgentRequested(s) {
		l, err := ssh.NewAgentListener()
		if err != nil {
			fmt.Printf(err.Error())
		}
		defer l.Close()
		go ssh.ForwardAgentConnections(l, s)
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", "SSH_AUTH_SOCK", l.Addr().String()))
	}

	cmd.Env = append(cmd.Env, fmt.Sprintf("TERM=%s", ptyReq.Term))
	f, err := pty.Start(cmd)
	if err != nil {
		panic(err)
	}

	sigs := make(chan ssh.Signal, 1)
	s.Signals(sigs)
	defer func() {
		s.Signals(nil)
		close(sigs)
	}()
	go func() {
		for {
			if sigs == nil && winCh == nil {
				return
			}

			select {
			case sig, ok := <-sigs:
				if !ok {
					sigs = nil
					continue
				}
				signal := osSignalFrom(sig)
				fmt.Println(signal.String())
			case win, ok := <-winCh:
				if !ok {
					winCh = nil
					continue
				}
				err = pty.Setsize(f, &pty.Winsize{
					Rows: uint16(win.Height),
					Cols: uint16(win.Width),
				})
				if err != nil {
					panic(err)
				}
			}
		}
	}()

	go func() {
		for win := range winCh {
			setWinsize(f, win.Width, win.Height)
		}
	}()
	go func() {
		io.Copy(f, s) // stdin
	}()
	io.Copy(s, f) // stdout

	log.Println("done")
}

func handleNonPty(s ssh.Session) {
	args := []string{}
	if len(s.Command()) > 0 {
		args = append([]string{"-c"}, s.Command()...)
	}

	cmd := exec.Command(getShell(), args...)

	if ssh.AgentRequested(s) {
		l, err := ssh.NewAgentListener()
		if err != nil {
			fmt.Printf(err.Error())
		}
		defer l.Close()
		go ssh.ForwardAgentConnections(l, s)
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", "SSH_AUTH_SOCK", l.Addr().String()))
	}

	cmd.Stdout = s
	cmd.Stderr = s.Stderr()
	// This blocks forever until stdin is received if we don't
	// use StdinPipe. It's unknown what causes this.
	stdinPipe, err := cmd.StdinPipe()
	if err != nil {
		log.Println(err)
		return
	}
	go func() {
		_, err := io.Copy(stdinPipe, s)
		if err != nil {
			log.Println(err)
		}
		_ = stdinPipe.Close()
	}()
	err = cmd.Start()
	if err != nil {
		log.Println(err)
		return
	}
	sigs := make(chan ssh.Signal, 1)
	s.Signals(sigs)
	defer func() {
		s.Signals(nil)
		close(sigs)
	}()
	go func() {
		for sig := range sigs {
			signal := osSignalFrom(sig)
			cmd.Process.Signal(signal)
		}
	}()
	cmd.Wait()
}

func osSignalFrom(sig ssh.Signal) os.Signal {
	switch sig {
	case ssh.SIGABRT:
		return unix.SIGABRT
	case ssh.SIGALRM:
		return unix.SIGALRM
	case ssh.SIGFPE:
		return unix.SIGFPE
	case ssh.SIGHUP:
		return unix.SIGHUP
	case ssh.SIGILL:
		return unix.SIGILL
	case ssh.SIGINT:
		return unix.SIGINT
	case ssh.SIGKILL:
		return unix.SIGKILL
	case ssh.SIGPIPE:
		return unix.SIGPIPE
	case ssh.SIGQUIT:
		return unix.SIGQUIT
	case ssh.SIGSEGV:
		return unix.SIGSEGV
	case ssh.SIGTERM:
		return unix.SIGTERM
	case ssh.SIGUSR1:
		return unix.SIGUSR1
	case ssh.SIGUSR2:
		return unix.SIGUSR2

	// Unhandled, use sane fallback.
	default:
		return unix.SIGKILL
	}
}

func getShell() string {
	shellEnv, shellSet := os.LookupEnv("SHELL")

	if shellSet {
		return shellEnv
	}

	return "sh"
}

func sftpHandler(sess ssh.Session) {
	debugStream := io.Discard
	serverOptions := []sftp.ServerOption{
		sftp.WithDebug(debugStream),
	}
	server, err := sftp.NewServer(
		sess,
		serverOptions...,
	)
	if err != nil {
		log.Printf("sftp server init error: %s\n", err)
		return
	}
	if err := server.Serve(); err == io.EOF {
		server.Close()
		fmt.Println("sftp client exited session.")
	} else if err != nil {
		fmt.Println("sftp server completed with error:", err)
	}
}
