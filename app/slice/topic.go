package slice

import (
	"sync"

	"github.com/genzai-io/sliced/proto/store"
	"github.com/genzai-io/sliced/common/btrdb"
)

var (
	TblTopic = &tblTopic{
		Table: newTable(
			"t",
			func() Serializable { return &store.Topic{} },
		),
	}
)

type tblTopic struct {
	*btrdb.Table

	byId map[int64]*store.Topic
	byName map[string]*store.Topic
}

//
type Topic struct {
	sync.RWMutex

	model store.Topic
}

func newTopic(model *store.Topic) *Topic {
	topic := &Topic{
		model: *model,
	}

	return topic
}

// Roll segment file
func (t *Topic) roll() {
	// Create the future segment
}
