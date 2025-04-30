package plugins

func LoadPlugin(pluginPath string) (PolicyFunction, error) {
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "POLICY_ENGINE",
			MagicCookieValue: "aws-policy-engine",
		},
		Plugins: map[string]plugin.Plugin{
			"policy_function": &PolicyFunctionPlugin{},
		},
		Cmd: exec.Command(pluginPath),
	})

	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	raw, err := rpcClient.Dispense("policy_function")
	return raw.(PolicyFunction), nil
}
