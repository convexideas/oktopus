package registry

import (
	"fmt"
	"regexp"

	"github.com/go-playground/validator/v10"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/providers/rawbytes"
	"github.com/knadh/koanf/v2"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
	validate.RegisterValidation("kebab", validateKebab)
}

var kebabRe = regexp.MustCompile(`^[a-z][a-z0-9-]*$`)

func validateKebab(fl validator.FieldLevel) bool {
	return kebabRe.MatchString(fl.Field().String())
}

// Metadata holds the manifest envelope fields.
type Metadata struct {
	Kind        string
	Name        string
	Version     string
	Description string
	Author      string
	Tags        []string
}

// LoadPersonaFile reads and validates a persona manifest YAML file.
func LoadPersonaFile(path string) (*Metadata, *PersonaSpec, error) {
	k := koanf.New(".")
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return nil, nil, fmt.Errorf("load manifest %s: %w", path, err)
	}
	return unmarshalPersona(k)
}

// ParsePersona validates persona manifest from raw YAML bytes.
func ParsePersona(data []byte) (*Metadata, *PersonaSpec, error) {
	k := koanf.New(".")
	if err := k.Load(rawbytes.Provider(data), yaml.Parser()); err != nil {
		return nil, nil, fmt.Errorf("parse manifest: %w", err)
	}
	return unmarshalPersona(k)
}

// --- unexported deserialization types ---

type manifest struct {
	Kind        string   `koanf:"kind" validate:"required,eq=persona"`
	Name        string   `koanf:"name" validate:"required,kebab"`
	Version     string   `koanf:"version" validate:"required"`
	Description string   `koanf:"description" validate:"required"`
	Author      string   `koanf:"author"`
	Tags        []string `koanf:"tags"`
	Persona     specYAML `koanf:"persona"`
}

type specYAML struct {
	SystemPrompt string                `koanf:"system_prompt" validate:"required"`
	PromptMode   string                `koanf:"prompt_mode" validate:"omitempty,oneof=append replace"`
	Skills       []string              `koanf:"skills"`
	OutputFormat string                `koanf:"output_format"`
	Native       map[string]nativeYAML `koanf:"native"`
}

type nativeYAML struct {
	Files    map[string]string `koanf:"files"`
	Flags    []string          `koanf:"flags"`
	Settings map[string]any    `koanf:"settings"`
}

func unmarshalPersona(k *koanf.Koanf) (*Metadata, *PersonaSpec, error) {
	var m manifest
	if err := k.Unmarshal("", &m); err != nil {
		return nil, nil, fmt.Errorf("unmarshal manifest: %w", err)
	}
	if err := validate.Struct(m); err != nil {
		return nil, nil, fmt.Errorf("validate manifest: %w", err)
	}

	meta := &Metadata{
		Kind:        m.Kind,
		Name:        m.Name,
		Version:     m.Version,
		Description: m.Description,
		Author:      m.Author,
		Tags:        m.Tags,
	}

	native := make(map[string]NativeConfig, len(m.Persona.Native))
	for name, nc := range m.Persona.Native {
		native[name] = NativeConfig{
			Files:    nc.Files,
			Flags:    nc.Flags,
			Settings: nc.Settings,
		}
	}

	spec := &PersonaSpec{
		SystemPrompt: m.Persona.SystemPrompt,
		PromptMode:   m.Persona.PromptMode,
		Skills:       m.Persona.Skills,
		OutputFormat:  m.Persona.OutputFormat,
		Native:       native,
	}

	return meta, spec, nil
}
