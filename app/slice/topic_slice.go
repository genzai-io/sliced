package slice

// Represents a Slice of a Topic. Each Slice is essentially
// a completely independent structure from all other slices,
// but shares the same ID and other meta-data.
type TopicSlice struct {
	parent *Topic
	slice  *Service
	key    string

	// The root or default partition
	root *TopicPartition
	// "named" topics within a slice are all independent partitions
	named map[string]*TopicPartition
}
