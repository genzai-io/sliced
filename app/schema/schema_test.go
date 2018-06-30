package schema

import (
	"fmt"
	"testing"

	"github.com/genzai-io/sliced/proto/schema"
	"github.com/gogo/protobuf/protoc-gen-gogo/descriptor"
)

func TestNewSchema(t *testing.T) {
	s := &schema.Schema{}
	fd, md := descriptor.ForMessage(s)

	fmt.Println(fd)
	_ = fd
	_ = md
}
