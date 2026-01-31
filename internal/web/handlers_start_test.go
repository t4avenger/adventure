package web

import (
	"testing"
)

func TestAllowedAvatar(t *testing.T) {
	for _, id := range AvatarOptions {
		if !allowedAvatar(id) {
			t.Errorf("allowedAvatar(%q) = false, want true", id)
		}
	}
	if allowedAvatar("") {
		t.Error("allowedAvatar(\"\") = true, want false")
	}
	if allowedAvatar("invalid") {
		t.Error("allowedAvatar(\"invalid\") = true, want false")
	}
	if allowedAvatar("male_youngx") {
		t.Error("allowedAvatar(\"male_youngx\") = true, want false")
	}
}
