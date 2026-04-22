package tests

import (
	"testing"

	"github.com/wesleiramos/mockpay/internal/util"
)

func TestSignPayload(t *testing.T) {
	secret := "test_secret"
	payload := []byte(`{"type":"billing.approved"}`)

	sig := util.SignPayload(secret, payload)

	if len(sig) < 10 {
		t.Errorf("signature too short: %s", sig)
	}

	if sig[:7] != "sha256=" {
		t.Errorf("expected sha256= prefix, got %s", sig[:7])
	}

	sig2 := util.SignPayload(secret, payload)
	if sig != sig2 {
		t.Error("expected same signature for same input")
	}

	sig3 := util.SignPayload("other_secret", payload)
	if sig == sig3 {
		t.Error("expected different signatures for different secrets")
	}
}

func TestNewID(t *testing.T) {
	id1 := util.NewID()
	id2 := util.NewID()

	if id1 == id2 {
		t.Error("expected different IDs")
	}

	if len(id1) < 16 {
		t.Errorf("ID too short: %s", id1)
	}
}

func TestNewIDsAreUnique(t *testing.T) {
	ids := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := util.NewID()
		if ids[id] {
			t.Error("generated duplicate ID")
		}
		ids[id] = true
	}
}
