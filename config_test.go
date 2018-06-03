package moved

import (
	"testing"
	"time"
)

func TestConfig(t *testing.T) {
	if err := readConfig(); err != nil {
		t.Fatal(err)
	}

	time.Sleep(time.Second * 100000)
}