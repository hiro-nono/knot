package ai

import (
	"strings"
	"testing"

	"knot-api/internal/domain"
)

func TestBuildMessages(t *testing.T) {
	got := BuildMessages("system prompt", "user input")

	if len(got) != 2 {
		t.Fatalf("len(got) = %d, want 2", len(got))
	}
	if got[0].Role != RoleSystem || got[0].Content != "system prompt" {
		t.Errorf("got[0] = %+v, want {Role: system, Content: system prompt}", got[0])
	}
	if got[1].Role != RoleUser || got[1].Content != "user input" {
		t.Errorf("got[1] = %+v, want {Role: user, Content: user input}", got[1])
	}
}

func TestBuildSystemPrompt(t *testing.T) {
	got := BuildSystemPrompt()

	for _, want := range []string{"identity", "fact", "schedule", "condition", "interaction", "confirmed", "undecided", "unknown", "question", "radio", "check", "text"} {
		if !strings.Contains(got, want) {
			t.Errorf("BuildSystemPrompt() does not contain %q:\n%s", want, got)
		}
	}
}

func TestBuildKnownKeysPrompt_Empty(t *testing.T) {
	if got := BuildKnownKeysPrompt(nil); got != "" {
		t.Errorf("BuildKnownKeysPrompt(nil) = %q, want empty string", got)
	}
	if got := BuildKnownKeysPrompt(map[domain.SourceType][]string{}); got != "" {
		t.Errorf("BuildKnownKeysPrompt({}) = %q, want empty string", got)
	}
}

func TestBuildKnownKeysPrompt_WithKeys(t *testing.T) {
	got := BuildKnownKeysPrompt(map[domain.SourceType][]string{
		domain.SourceTypeSchedule: {"departure_date", "start_time"},
		domain.SourceTypeIdentity: {"organizer"},
	})

	for _, want := range []string{"schedule", "departure_date", "start_time", "identity", "organizer", "再利用"} {
		if !strings.Contains(got, want) {
			t.Errorf("BuildKnownKeysPrompt() does not contain %q:\n%s", want, got)
		}
	}
}

func TestBuildKnownPreferenceKeysPrompt_Empty(t *testing.T) {
	if got := BuildKnownPreferenceKeysPrompt(nil); got != "" {
		t.Errorf("BuildKnownPreferenceKeysPrompt(nil) = %q, want empty string", got)
	}
	if got := BuildKnownPreferenceKeysPrompt([]string{}); got != "" {
		t.Errorf("BuildKnownPreferenceKeysPrompt([]) = %q, want empty string", got)
	}
}

func TestBuildKnownPreferenceKeysPrompt_WithKeys(t *testing.T) {
	got := BuildKnownPreferenceKeysPrompt([]string{"font_size", "language", "visual_style"})

	for _, want := range []string{"font_size", "language", "visual_style", "再利用"} {
		if !strings.Contains(got, want) {
			t.Errorf("BuildKnownPreferenceKeysPrompt() does not contain %q:\n%s", want, got)
		}
	}
}
