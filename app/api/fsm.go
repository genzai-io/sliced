package api

//
type ApplyRecord interface {
	Count() int32

	Commands() []Command
}

// A single command in an ApplyRecord to apply to the FSM
type ApplyCommand interface {
}
