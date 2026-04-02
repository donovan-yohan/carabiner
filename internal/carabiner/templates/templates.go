package templates

import (
	"fmt"
)

// Template represents a project template with configuration files
type Template struct {
	Name        string            // Template name (e.g., "react-typescript", "go")
	EnforceYAML string            // Content for enforce.yaml
	ConfigFiles map[string]string // Additional config files (filename -> content)
}

// GetTemplate returns a template by name
func GetTemplate(name string) (*Template, error) {
	switch name {
	case "react-typescript":
		return getReactTypescriptTemplate(), nil
	case "go":
		return getGoTemplate(), nil
	case "svelte-kit":
		return getSvelteKitTemplate(), nil
	case "svelte-vite":
		return getSvelteViteTemplate(), nil
	case "svelte":
		return getSvelteKitTemplate(), nil
	default:
		return nil, fmt.Errorf("unknown template: %s", name)
	}
}

// ListTemplates returns all available template names
func ListTemplates() []string {
	return []string{
		"go",
		"react-typescript",
		"svelte-kit",
		"svelte-vite",
	}
}
