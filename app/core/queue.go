package core

type Queue struct {
	request *Topic
	reply   *Topic
	error   *Topic
}

type QueueSlice struct {
	request *TopicPartition
	reply   *TopicPartition
	error   *TopicPartition
}
