package cmd_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/app/cmd"
	"github.com/movemedical/server-agent/common/redcon"
)

var raftJoin = &cmd.RaftJoin{}

func TestRaftJoin_Registry(t *testing.T) {
	_, ok := api.Commands[raftJoin.Name()]
	if !ok {
		panic(raftJoin.Name() + " not registered")
	}
	_, ok = api.Commands[strings.ToUpper(raftJoin.Name())]
	if !ok {
		panic(strings.ToUpper(raftJoin.Name()) + " not registered")
	}
	_, ok = api.Commands[strings.ToLower(raftJoin.Name())]
	if !ok {
		panic(strings.ToLower(raftJoin.Name()) + " not registered")
	}
}

func TestRaftJoin_Marshall(t *testing.T) {
	c, ok := api.Commands[raftJoin.Name()]
	if !ok {
		panic(raftJoin.Name() + " not registered")
	}
	if c == nil {
		t.Fatal(errors.New(raftJoin.Name() + " registered nil command"))
	}


	for _, cc := range []*cmd.RaftJoin{
		{
			ID:      api.GlobalRaftID,
			Address: "127.0.0.1:9001",
			Voter:   false,
		},
		{
			ID:      api.GlobalRaftID,
			Address: "127.0.0.1:9001",
			Voter:   true,
		},
		{
			ID: api.RaftID{
				DatabaseID: 0,
				SliceID:    0,
			},
			Address: "127.0.0.1:9001",
			Voter:   false,
		},
		{
			ID: api.RaftID{
				DatabaseID: 1,
				SliceID:    1,
			},
			Address: "127.0.0.1:9001",
			Voter:   false,
		},
		{
			ID: api.RaftID{
				DatabaseID: 1,
				SliceID:    1,
			},
			Address: "127.0.0.1:9001",
			Voter:   true,
		},
		{
			ID: api.RaftID{
				DatabaseID: 3,
				SliceID:    5,
			},
			Address: "127.0.0.1:9001",
			Voter:   true,
		},
	} {
		testMarshalRaftJoin(t, cc)
	}
}

func testMarshalRaftJoin(t *testing.T, c *cmd.RaftJoin) {
	args, _, err := redcon.ParseCommand(c.Marshal(nil))
	if err != nil {
		t.Fatal(err)
	}

	cmd2 := c.Parse(args)

	if !compareRaftJoin(c, cmd2.(*cmd.RaftJoin)) {
		fmt.Println(fmt.Sprintf("%s\n!=\n%s", c, cmd2))
		t.Fatal(errors.New("compare failed"))
	}
}

func compareRaftJoin(r1 *cmd.RaftJoin, r2 *cmd.RaftJoin) bool {
	if r1.ID.DatabaseID != r2.ID.DatabaseID {
		return false
	}
	if r1.ID.SliceID != r2.ID.SliceID {
		return false
	}
	if r1.Address != r2.Address {
		return false
	}
	if r1.Voter != r2.Voter {
		return false
	}

	return true
}
