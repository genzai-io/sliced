package command

import "github.com/slice-d/genzai/app/api"

func init() {
	api.Handler = &handler{}
}

type handler struct {
}
