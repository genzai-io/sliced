package main

import (
	"fmt"

	"github.com/gbrlsnchs/radix"
	"github.com/slice-d/genzai/item"
)

func main() {
	// Example from https://upload.wikimedia.org/wikipedia/commons/a/ae/Patricia_trie.svg.
	t := radix.New("Example")

	for i := uint64(0); i < 10; i++ {
		key := item.UInt64ToString(i)
		//key := strconv.Itoa(i)
		t.Add(key, i)
	}

	//t.Add("romane", 1)
	t.Add("romanus", 2)
	//t.Add("romulus", 3)
	//t.Add("rubens", 4)
	//t.Add("ruber", 5)
	//t.Add("rubicon", 6)
	//t.Add("rubicundus", 7)
	t.Sort(radix.AscLabelSort)

	err := t.Debug()

	if err != nil {
		// ...
	}

	t.Sort(radix.DescLabelSort)

	err = t.Debug()

	if err != nil {
		// ...
	}

	t.Sort(radix.PrioritySort)

	err = t.Debug()

	if err != nil {
		// ...
	}

	n := t.Get("romanus")

	fmt.Println(n.Value)
}
