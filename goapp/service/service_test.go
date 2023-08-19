package service

import (
	"os"
	"testing"
)

func TestListenAddress(t *testing.T) {
	_ = os.Setenv("ADDRESS", "")
	_ = os.Setenv("PORT", "")
	if ListenAddress() != "localhost:8080" {
		t.Fail()
	}

	_ = os.Setenv("ADDRESS", "0.0.0.0:8080")
	_ = os.Setenv("PORT", "")
	if ListenAddress() != "0.0.0.0:8080" {
		t.Fail()
	}

	_ = os.Setenv("ADDRESS", "")
	_ = os.Setenv("PORT", "9999")
	if ListenAddress() != "localhost:9999" {
		t.Fail()
	}
}
