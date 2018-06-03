package cmd

import "github.com/genzai-io/sliced/app/api"

func init() {
	api.Handler = &handler{}
}

type handler struct {
}
