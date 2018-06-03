package path

import (
	"fmt"
	"path/filepath"

	"github.com/slice-d/genzai/proto/store"
)

func SegmentPath(segment *store.Segment) string {
	if segment == nil || segment.Path == nil {
		return ""
	}

	return filepath.Join(
		segment.Path.Drive, // /drive
		"s", // slices folder
		fmt.Sprintf("%d", segment.Slice), // Slice index
		"t", // Topics folder
		fmt.Sprintf("%d", segment.TopicID), // Topic id
		fmt.Sprintf("%d.s", segment.Id), // Segment sequence
	)
}
