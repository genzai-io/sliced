package core

import "testing"

func TestService_OnStart(t *testing.T) {
	err := Instance.Start()
	if err != nil {
		t.Fatal(err)
	}

	err = Instance.Stop()
	if err != nil {
		t.Fatal(err)
	}
}
