package tmux

import "testing"

func TestSanitizeSessionName(t *testing.T) {
	cases := map[string]string{
		"build":           "build",
		"my build.server": "my-build-server",
		"api:v2":          "api-v2",
		"-d":              "d",
		"--rf":            "rf",
		"a\nb":            "a-b",
		"   ":             "",
		"":                "",
		"  spaced name ":  "spaced-name",
	}
	for in, want := range cases {
		if got := SanitizeSessionName(in); got != want {
			t.Errorf("SanitizeSessionName(%q) = %q, want %q", in, got, want)
		}
	}
}
