package timersocket_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWebsocket(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	connLength := 100
	conns := make([]*WsConn, 0, connLength)

	for i := 0; i < connLength; i++ {
		conn := NewConn(t, ctx, server, int64(i))
		conns = append(conns, conn)
		go conn.Listen(t, ctx)
	}

	for _, conn := range conns {
		err := conn.Close()
		require.NoError(t, err)
	}
}
