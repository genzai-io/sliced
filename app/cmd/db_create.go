package cmd

import (
	"strings"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/redcon"
)

func init() { api.Register(api.CreateDatabase, &CreateDatabase{}) }

// Demotes a Voting member to a Non-Voting member.
type CreateDatabase struct {
	Name string

	raft api.RaftService
}

func (c *CreateDatabase) IsChange() bool { return false }
func (c *CreateDatabase) IsAsync() bool  { return false }

func (c *CreateDatabase) Marshal(buf []byte) []byte {
	buf = redcon.AppendArray(buf, 2)
	buf = redcon.AppendBulkString(buf, api.CreateDatabase)
	buf = redcon.AppendBulkString(buf, c.Name)
	return buf
}

func (c *CreateDatabase) Parse(ctx *Context) api.Command {
	cmd := &CreateDatabase{}

	switch len(ctx.Args) {
	default:
		ctx.Err("invalid params")
		return cmd

	case 2:
		cmd.Name = strings.TrimSpace(string(ctx.Args[1]))

		if len(cmd.Name) == 0 {
			ctx.Err("name not set")
			return cmd
		}

		return cmd
	}
	return cmd
}

func (c *CreateDatabase) Handle(ctx *Context) {
	// Create in the store.

	ctx.OK()
}

func (c *CreateDatabase) Apply(ctx *Context) {}
