package core

import (
	"github.com/genzai-io/sliced/app/ring"
	"github.com/genzai-io/sliced/common/service"
)

// Manages all locally joined slices
type Master struct {
	service.BaseService

	table [ring.Slots]*SliceService
}


