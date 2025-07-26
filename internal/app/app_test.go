package app

import "testing"

func TestAppVersion(t *testing.T) {
	app := New()

	if app.Version != "0.1.5" {
		t.Errorf("Expected version 0.1.5, got %s", app.Version)
	}

	if app.Name != "logbasset" {
		t.Errorf("Expected name logbasset, got %s", app.Name)
	}

	expectedFullVersion := "logbasset version 0.1.5"
	if app.GetFullVersion() != expectedFullVersion {
		t.Errorf("Expected full version '%s', got '%s'", expectedFullVersion, app.GetFullVersion())
	}
}

func TestAppConstants(t *testing.T) {
	if Version != "0.1.5" {
		t.Errorf("Expected Version constant to be 0.1.5, got %s", Version)
	}

	if Name != "logbasset" {
		t.Errorf("Expected Name constant to be logbasset, got %s", Name)
	}

	if Author != "Andrea Grandi" {
		t.Errorf("Expected Author constant to be Andrea Grandi, got %s", Author)
	}

	if License != "Apache-2.0" {
		t.Errorf("Expected License constant to be Apache-2.0, got %s", License)
	}
}
