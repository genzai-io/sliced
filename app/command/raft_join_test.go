package command_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/slice-d/genzai/app/api"
	"github.com/slice-d/genzai/app/command"
)

func TestRaftJoin_Registry(t *testing.T) {
	_, ok := api.Commands[api.RaftJoinName]
	if !ok {
		panic(api.RaftJoinName + " not registered")
	}
	_, ok = api.Commands[strings.ToUpper(api.RaftJoinName)]
	if !ok {
		panic(strings.ToUpper(api.RaftJoinName) + " not registered")
	}
	_, ok = api.Commands[strings.ToLower(api.RaftJoinName)]
	if !ok {
		panic(strings.ToLower(api.RaftJoinName) + " not registered")
	}
}

func TestRaftJoin_Marshall(t *testing.T) {
	cmd, ok := api.Commands[api.RaftJoinName]
	if !ok {
		panic(api.RaftJoinName + " not registered")
	}
	if cmd == nil {
		t.Fatal(errors.New(api.RaftJoinName + " registered nil command"))
	}

	for _, cmd := range []*command.RaftJoin{
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
				Schema: 0,
				Slice:  0,
			},
			Address: "127.0.0.1:9001",
			Voter:   false,
		},
		{
			ID: api.RaftID{
				Schema: 1,
				Slice:  1,
			},
			Address: "127.0.0.1:9001",
			Voter:   false,
		},
		{
			ID: api.RaftID{
				Schema: 1,
				Slice:  1,
			},
			Address: "127.0.0.1:9001",
			Voter:   true,
		},
		{
			ID: api.RaftID{
				Schema: 3,
				Slice:  5,
			},
			Address: "127.0.0.1:9001",
			Voter:   true,
		},
	} {
		testMarshalRaftJoin(t, cmd)
	}
}

func testMarshalRaftJoin(t *testing.T, cmd *command.RaftJoin) {
	buf := cmd.Marshal(nil)
	ctx := createContext(t, buf)

	cmd2 := cmd.Parse(ctx)

	if !compareRaftJoin(cmd, cmd2.(*command.RaftJoin)) {
		fmt.Println(fmt.Sprintf("%s\n!=\n%s", cmd, cmd2))
		t.Fatal(errors.New("compare failed"))
	}
}

func compareRaftJoin(r1 *command.RaftJoin, r2 *command.RaftJoin) bool {
	if r1.ID.Schema != r2.ID.Schema {
		return false
	}
	if r1.ID.Slice != r2.ID.Slice {
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
