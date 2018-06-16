package document

import "github.com/genzai-io/sliced/app/table"

type Projector interface {
	Project(b []byte) table.Key

	ProjectString(s string) table.Key
}

var NilProjector = nilProjector{}

type nilProjector struct{}

func (n nilProjector) Project(b []byte) table.Key {
	return table.Nil
}

func (n nilProjector) ProjectString(s string) table.Key {
	return table.Nil
}
