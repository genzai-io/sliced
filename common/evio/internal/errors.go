package internal

import "fmt"

var (
	ErrPanicked = func(reason interface{}) error {
		//switch reason.(type) {
		//case string:
		//
		//case error:
		//
		//}
		return fmt.Errorf("panic: %s", reason)
	}
)
