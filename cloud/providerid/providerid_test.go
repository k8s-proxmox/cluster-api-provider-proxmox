package providerid

import (
	"testing"
)

func TestNew(t *testing.T) {
	uuid := ""
	providerID, err := New(uuid)
	if err == nil {
		t.Errorf("err should not be nil. providerID=%s", providerID)
	}

	uuid = "asdf"
	providerID, err = New(uuid)
	if err != nil {
		t.Errorf("failed to create providerID: %v", err)
	}
}

func TestUUID(t *testing.T) {
	pid := providerID{
		uuid: "asdf",
	}
	uuid := pid.UUID()
	if uuid != "asdf" {
		t.Errorf("should be asdf")
	}
}

func TestString(t *testing.T) {
	pid := providerID{
		uuid: "asdf",
	}
	uuid := pid.String()
	if uuid != Prefix+"asdf" {
		t.Errorf("should be %sasdf", Prefix)
	}
}
