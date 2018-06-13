package cmd_test

import (
	"testing"

	"github.com/genzai-io/sliced/app/api"
	"github.com/genzai-io/sliced/common/resp"
)

func createContext(t *testing.T, buf []byte) *api.Context {
	args, _, err := resp.ParseCommand(buf)
	if err != nil {
		t.Fatal(err)
	}

	ctx := &api.Context{
		Args: args,
	}
	return ctx
}
