package ring

import (
	"fmt"
	"testing"
)

func TestDiff(t *testing.T) {
	ring2 := Balanced(4)
	fmt.Println(ring2)
	fmt.Println()

	ring := Balanced(10)
	fmt.Println(ring)

	changes, err := ring.Migrate(ring2)

	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(RebalanceString(changes))

	changes, err = ring2.Migrate(ring)

	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(RebalanceString(changes))
}
