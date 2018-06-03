package core

import "github.com/slice-d/genzai/proto/store"

type Slice struct {
	model  *store.Slice
	owned  bool
	nodes  []*Node

	service *SliceService
}

func newSlice(model *store.Slice, owned bool, nodes []*Node) *Slice {
	s := &Slice{
		model: model,
		owned: owned,
		nodes: nodes,
	}

	if owned {
		//s.service = newSliceService()
	}

	return s
}

func (s *Slice) startService() error {
	return nil
}

