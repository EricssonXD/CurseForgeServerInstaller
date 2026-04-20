package errors

import "testing"

func TestIsUserFacing(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"UserFacingError", NewUserFacingError("oops"), true},
		{"UserFacingErrorf", NewUserFacingErrorf("oops %d", 42), true},
		{"MissingApiKeyError", NewMissingApiKeyError("no key"), true},
		{"InvalidApiKeyError", NewInvalidApiKeyError("bad key"), true},
		{"generic error", &McServerError{Message: "base"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsUserFacing(tt.err); got != tt.want {
				t.Errorf("IsUserFacing() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorMessages(t *testing.T) {
	if msg := NewUserFacingErrorf("hello %s", "world").Error(); msg != "hello world" {
		t.Errorf("got %q", msg)
	}
}
