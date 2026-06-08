package api

import (
	"github.com/charmbracelet/ssh"
	"github.com/spf13/cobra"
)

// RecvStream is the minimal interface implemented by a gRPC server-streaming
// client (e.g. grpc.ServerStreamingClient[T]). It is satisfied by the streams
// returned by SubscribeServerLogs and SubscribeNotifications.
type RecvStream[T any] interface {
	Recv() (*T, error)
}

// StreamMessages consumes a gRPC server stream until the SSH client sends a
// break (Ctrl-C) or the stream returns an error. Each received message is
// passed to the handler for rendering. This is shared by the "livelog" and
// "notifications" commands, which both follow the same break/recv loop.
func StreamMessages[T any](cobraCmd *cobra.Command, stream RecvStream[T], handler func(msg *T)) error {
	sshSession := cobraCmd.Context().Value("ssh").(*ssh.Session)
	cobraCmd.SetOut(*sshSession)

	connectionClosed := false
	go func() {
		breakChan := make(chan bool)
		(*sshSession).Break(breakChan)
		<-breakChan
		connectionClosed = true
	}()

	cobraCmd.Printf("\033[33mWaiting for new notifications ...\033[0m\n")
	for {
		if connectionClosed {
			break
		}

		msg, err := stream.Recv() // Blocking
		if err != nil {
			return err
		}

		handler(msg)
	}

	return nil
}
