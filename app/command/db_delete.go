package command

import (
	"strings"

	"github.com/slice-d/genzai/app/api"
	"github.com/slice-d/genzai/common/redcon"
)

func init() { api.Register(api.CreateDatabase, &CreateDatabase{}) }

// Demotes a Voting member to a Non-Voting member.
type DeleteDatabase struct {
	Name string

	raft api.RaftService
}

func (c *DeleteDatabase) IsChange() bool { return false }
func (c *DeleteDatabase) IsAsync() bool  { return true }

func (c *DeleteDatabase) Marshal(buf []byte) []byte {
	buf = redcon.AppendArray(buf, 2)
	buf = redcon.AppendBulkString(buf, api.CreateDatabase)
	buf = redcon.AppendBulkString(buf, c.Name)
	return buf
}

func (c *DeleteDatabase) Parse(ctx *Context) api.Command {
	cmd := &DeleteDatabase{}

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

func (c *DeleteDatabase) Handle(ctx *Context) {
	// Create an apply

	ctx.OK()
}

func (c *DeleteDatabase) Apply(ctx *Context) {
	// Create in the store.
}
