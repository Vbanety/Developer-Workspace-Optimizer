package core

// Registry holds the modules available for the current OS, keyed by Name().
type Registry struct {
	modules map[string]Module
	order   []string // preserves registration order for stable report output
}

func NewRegistry() *Registry {
	return &Registry{modules: make(map[string]Module)}
}

// Register adds a module to the registry. Panics on duplicate names —
// that's a programming error caught at startup, not a runtime condition.
func (r *Registry) Register(m Module) {
	name := m.Name()
	if _, exists := r.modules[name]; exists {
		panic("core: module already registered: " + name)
	}
	r.modules[name] = m
	r.order = append(r.order, name)
}

// Get returns the module for a given name, or nil if not registered.
func (r *Registry) Get(name string) Module {
	return r.modules[name]
}

// All returns modules in registration order.
func (r *Registry) All() []Module {
	out := make([]Module, 0, len(r.order))
	for _, name := range r.order {
		out = append(out, r.modules[name])
	}
	return out
}
