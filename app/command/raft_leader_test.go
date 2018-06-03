package command_test

import (
	"strings"
	"testing"

	"github.com/slice-d/genzai/app/api"
)

func TestRaftLeader_Registry(t *testing.T) {
	_, ok := api.Commands[api.RaftLeaderName]
	if !ok {
		panic(api.RaftLeaderName + " not registered")
	}
	_, ok = api.Commands[strings.ToUpper(api.RaftLeaderName)]
	if !ok {
		panic(strings.ToUpper(api.RaftLeaderName) + " not registered")
	}
	_, ok = api.Commands[strings.ToLower(api.RaftLeaderName)]
	if !ok {
		panic(strings.ToLower(api.RaftLeaderName) + " not registered")
	}
}
