package voice

import "testing"

func TestValidateFallsBackToDictate(t *testing.T) {
	sessions := []string{"sapiologie-1", "sapiologie-2", "sapiologie-3"}

	got := Validate(Action{Type: "nonsense"}, sessions, "hello world")
	if got.Type != "dictate" || got.Text != "hello world" {
		t.Fatalf("expected dictate fallback, got %+v", got)
	}
}

func TestValidateDictateUsesTranscriptVerbatim(t *testing.T) {
	// The router must never reword dictation: even when it returns its own
	// (possibly rewritten) text, Validate replaces it with the exact transcript.
	got := Validate(Action{Type: "dictate", Text: "a reworded version"}, nil, "the exact words I said")
	if got.Type != "dictate" || got.Text != "the exact words I said" {
		t.Fatalf("expected verbatim transcript, got %+v", got)
	}
}

func TestValidateKeyAllowlist(t *testing.T) {
	got := Validate(Action{Type: "key", Name: "ctrl_c"}, nil, "t")
	if got.Type != "key" || got.Name != "ctrl_c" {
		t.Fatalf("expected key ctrl_c, got %+v", got)
	}

	got = Validate(Action{Type: "key", Name: "rm_rf"}, nil, "t")
	if got.Type != "dictate" {
		t.Fatalf("expected dictate for disallowed key, got %+v", got)
	}
}

func TestValidateSwitchSession(t *testing.T) {
	sessions := []string{"sapiologie-1", "sapiologie-2"}

	got := Validate(Action{Type: "switch_session", Target: "sapiologie-2"}, sessions, "t")
	if got.Type != "switch_session" || got.Target != "sapiologie-2" {
		t.Fatalf("expected switch to sapiologie-2, got %+v", got)
	}

	got = Validate(Action{Type: "switch_session", Target: "ghost"}, sessions, "t")
	if got.Type != "dictate" {
		t.Fatalf("expected dictate for unknown session, got %+v", got)
	}
}

func TestValidateScrollDefaults(t *testing.T) {
	got := Validate(Action{Type: "scroll", Dir: "sideways", Amount: -5}, nil, "t")
	if got.Type != "scroll" || got.Dir != "up" || got.Amount != 3 {
		t.Fatalf("expected normalized scroll, got %+v", got)
	}
}

func TestValidateCopyDefaultsToScreen(t *testing.T) {
	got := Validate(Action{Type: "copy"}, nil, "t")
	if got.Type != "copy" || got.Scope != "screen" {
		t.Fatalf("expected copy screen, got %+v", got)
	}
}
