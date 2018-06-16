package slice

import (
	"github.com/genzai-io/sliced/app/node"
	"github.com/genzai-io/sliced/proto/store"
)

type Slice struct {
	model *store.Slice
	owned bool
	nodes []*node.Node

	service *Service
}

func newSlice(model *store.Slice, owned bool, nodes []*node.Node) *Slice {
	s := &Slice{
		model: model,
		owned: owned,
		nodes: nodes,
	}

	if owned {
		//s.service = newService()
	}

	return s
}

func (s *Slice) startService() error {
	return nil
}
