package cmd_test

import (
	"testing"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/redcon"
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
