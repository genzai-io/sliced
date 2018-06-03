package core

import "github.com/genzai-io/sliced/proto/store"

// A bucket represents a bucket in a cloud object storage system like S3.
type Bucket struct {
	model *store.Bucket
}

func newBucket() (*Bucket, error) {
	return nil, nil
}
