package registry_test

import (
	"testing"

	"github.com/convexideas/oktopus/internal/registry"
)

func TestLoadPersonaFile(t *testing.T) {
	meta, spec, err := registry.LoadPersonaFile("../../registry/personas/code-reviewer.yaml")
	if err != nil {
		t.Fatalf("LoadPersonaFile: %v", err)
	}

	if meta.Kind != "persona" {
		t.Errorf("kind = %q, want persona", meta.Kind)
	}
	if meta.Name != "code-reviewer" {
		t.Errorf("name = %q, want code-reviewer", meta.Name)
	}
	if meta.Version != "0.1.0" {
		t.Errorf("version = %q, want 0.1.0", meta.Version)
	}
	if spec.SystemPrompt == "" {
		t.Error("system_prompt is empty")
	}

	native, ok := spec.Native["pi"]
	if !ok {
		t.Fatal("missing native config for pi")
	}
	if len(native.Flags) == 0 {
		t.Error("pi native flags are empty")
	}
}

func TestParsePersonaInvalid(t *testing.T) {
	cases := []struct {
		name string
		yaml string
	}{
		{"missing kind", "name: test\nversion: '0.1.0'\ndescription: x\npersona:\n  system_prompt: hi"},
		{"bad kind", "kind: tool\nname: test\nversion: '0.1.0'\ndescription: x\npersona:\n  system_prompt: hi"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, _, err := registry.ParsePersona([]byte(tc.yaml))
			if err == nil {
				t.Error("expected validation error")
			}
		})
	}
}
