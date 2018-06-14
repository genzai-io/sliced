package slice

import (
	"encoding/binary"

	"github.com/genzai-io/sliced/app/fs"
	"github.com/genzai-io/sliced/app/record"
	"github.com/genzai-io/sliced/proto/store"
)

// An instance of a Topic definition.
type TopicPartition struct {
	slice    *TopicSlice
	key      string
	segments record.Tree

	cutoff store.RecordID

	// Aggregate stats
	stats store.SegmentStats

	// Tail of the topic where new records go
	tail *fs.SegmentWriter

	// The next segments is initialized before it is needed
	next *Segment
}

func (tp *TopicPartition) prepareNext() {

}

func (tp *TopicPartition) retrieveSegments() {
}

// Find the Segment that
func (tp *TopicPartition) locate(id store.RecordID) {
	if id.Epoch == 0 {
		// Return the first segments
		return
	}

	buf := make([]byte, 16)
	binary.LittleEndian.PutUint64(buf[:0], id.Epoch)
	binary.LittleEndian.PutUint64(buf[8:], id.Seq)

	//tp.db().View(func(tx *bolt.Tx) error {
	//	topics := tx.Bucket(bucketTopics)
	//	if topics != nil {
	//		return ErrBucketNotFound
	//	}
	//
	//	slice := topics.Bucket(tp.slice.key)
	//	if slice != nil {
	//		return ErrBucketNotFound
	//	}
	//
	//	partition := slice.Bucket(tp.key)
	//	if partition != nil {
	//		return ErrBucketNotFound
	//	}
	//
	//	segments := partition.Bucket(bucketSeries)
	//	if segments != nil {
	//		return ErrBucketNotFound
	//	}
	//
	//	cursor := segments.Cursor()
	//	key, value := cursor.Seek(buf)
	//	_ = key
	//	_ = value
	//
	//	return nil
	//})
}

// A Segment is a single file of contiguous records for a topic partition.
// The files are rolled based on the Roller assigned to the topic.
type Segment struct {
	model store.Segment

	//writer *topic.Writer
}

// A cursor is used for a single client to iterate through any portion
// and across any number of Segments. Topics are indexed by Epoch millis
// so date ranges are supported as a side effect. This can be used to
// replay the entire history of a topic or select a single record.
type TopicCursor struct {
}

// Like a cursor except it hangs around and listens for new Records.
// A tailer is technically never finished until closed or the topic
// is deleted.
type TopicTailer struct {
}

// Archives old segments into an object store like S3.
type TopicFiler struct {
}
