package plugins

type PolicyFunction interface {
	Execute(params map[string]interface{}) (interface{}, error)
	Metadata() PluginMetadata
}

type PluginMetadata struct {
	Name        string
	Description string
	Version     string
}
