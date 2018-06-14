package queue

import "github.com/genzai-io/sliced/app/fs"

type Queue struct {
	writer *fs.SegmentWriter
}

//type Queue struct {
//	request *Topic
//	reply   *Topic
//	error   *Topic
//}
//
//type QueueSlice struct {
//	request *TopicPartition
//	reply   *TopicPartition
//	error   *TopicPartition
//}

