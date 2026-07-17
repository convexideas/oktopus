package runtime

import "fmt"

// HarnessRegistry holds registered harness adapters.
type HarnessRegistry struct {
	adapters map[string]Harness
}

func NewHarnessRegistry() *HarnessRegistry {
	return &HarnessRegistry{adapters: make(map[string]Harness)}
}

func (r *HarnessRegistry) Register(h Harness) {
	r.adapters[h.Name()] = h
}

func (r *HarnessRegistry) Get(name string) (Harness, error) {
	h, ok := r.adapters[name]
	if !ok {
		names := make([]string, 0, len(r.adapters))
		for n := range r.adapters {
			names = append(names, n)
		}
		return nil, fmt.Errorf("harness not found: %s (available: %v)", name, names)
	}
	return h, nil
}

func (r *HarnessRegistry) List() []string {
	names := make([]string, 0, len(r.adapters))
	for n := range r.adapters {
		names = append(names, n)
	}
	return names
}
