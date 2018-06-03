package core

import (
	"github.com/slice-d/genzai/app/ring"
	"github.com/slice-d/genzai/common/service"
)

// Manages all locally joined slices
type Master struct {
	service.BaseService

	table [ring.Slots]*SliceService
}


