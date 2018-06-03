package command_test

import (
	"testing"

	"github.com/slice-d/genzai/app/api"
	"github.com/slice-d/genzai/common/redcon"
)

func createContext(t *testing.T, buf []byte) *api.Context {
	args, _, err := redcon.ParseCommand(buf)
	if err != nil {
		t.Fatal(err)
	}

	ctx := &api.Context{
		Args: args,
	}
	return ctx
}
