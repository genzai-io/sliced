package node

import (
	"time"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/redigo/redis"
)

// Inter-node communication / forwarding
type Transport interface {
	Get() Transport

	Send(command api.Command) api.CommandReply

	SendMany(commands ...api.Command) []api.CommandReply
}

type localTransport struct {
}

func (t *localTransport) Get() Transport {
	return t
}

func (t *localTransport) Send(command api.Command) api.CommandReply {
	command.Handle(nil)
	return nil
}

func (t *localTransport) SendMany(commands ...api.Command) []api.CommandReply {
	return nil
}

type remoteTransport struct {
	pool *redis.Pool
}

// Create a remote transport that uses the CmdConn as a means of communication
func newRemoteTransport(target string) *remoteTransport {
	return &remoteTransport{
		pool: &redis.Pool{
			//MaxIdle: 5, // figure 5 should suffice most clusters.
			//MaxActive:   25,
			IdleTimeout: time.Minute, //
			Wait:        false,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", target)
				if err != nil {
					return nil, err
				}
				return c, err
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				if time.Since(t) < time.Minute {
					return nil
				}
				//_, err := c.Do("PING")
				return nil
			},
		},
	}
}

func (t *remoteTransport) Get() Transport {
	conn := t.pool.Get()
	return &remoteTransportConn{
		transport: t,
		conn:      conn,
	}
}

func (t *remoteTransport) Send(command api.Command) api.CommandReply {
	conn := t.pool.Get()
	if conn == nil {
		return nil
	}

	var out []byte
	out = command.Marshal(out)

	//conn.Send()

	return nil
}

func (t *remoteTransport) SendMany(commands ...api.Command) []api.CommandReply {
	return nil
}

type remoteTransportConn struct {
	transport *remoteTransport
	conn      redis.Conn
}

func (t *remoteTransportConn) Get() Transport {
	return t.transport.Get()
}

func (t *remoteTransportConn) Send(command api.Command) api.CommandReply {
	//resp.ParseCommand()
	return nil
}

func (t *remoteTransportConn) SendMany(commands ...api.Command) []api.CommandReply {

	return nil
}
