package node

import (
	"time"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/redigo/redis"
)

// Inter-node communication / forwarding
type Transport interface {
	Send(command api.Command) []byte
}

type localTransport struct {
}

func (t *localTransport) Send(command api.Command) []byte {
	command.Handle(nil)
	return nil
}

type remoteTransport struct {
	pool *redis.Pool
}

func newRemoteTransport(target string) *remoteTransport {
	return &remoteTransport{
		pool: &redis.Pool{
			MaxIdle: 5, // figure 5 should suffice most clusters.
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
				_, err := c.Do("PING")
				return err
			},
		},
	}
}

func (t *remoteTransport) Get() redis.Conn {
	return t.pool.Get()
}

func (t *remoteTransport) Send(command api.Command) []byte {
	conn := t.pool.Get()
	if conn == nil {
		return nil
	}

	//resp.ParseCommand()
	return nil
}
