package main

type CustomValidator struct{}

func (c *CustomValidator) Execute(params map[string]interface{}) (interface{}, error) {
	// Custom logic here
	return params["input"], nil
}

func (c *CustomValidator) Metadata() plugins.PluginMetadata {
	return plugins.PluginMetadata{
		Name:        "custom-validator",
		Description: "Custom input validator",
		Version:     "1.0.0",
	}
}

// go build -o plugins/custom-validator.so -buildmode=plugin custom_plugin.go
// engine.LoadPlugins("plugins/")
