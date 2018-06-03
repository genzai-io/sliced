package core

import (
	"sync"

	"github.com/genzai-io/sliced/proto/store"
)

//
type Topic struct {
	sync.RWMutex

	key string // "t:{topicID}"
	// Model
	model *store.Topic

	// Stats
	mapped      int64
	mappedWrite int64
}

func newTopic(model *store.Topic) *Topic {
	topic := &Topic{
		model: model,
	}

	return topic
}

// Roll segment file
func (t *Topic) roll() {
	// Create the future segment
}
