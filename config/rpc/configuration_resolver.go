package rpc

import "context"

// ConfigurationResolver defines a means of resolving an RPC configuration.
type ConfigurationResolver interface {
	// ResolveConfiguration resolves an RPC configuration for the given chain name.
	// Returns true if a configuration was found; false if not.
	ResolveConfiguration(ctx context.Context, chainName string) (Configuration, bool, error)
}

// DefaultConfigurationResolver is a default implementation of ConfigurationResolver.
type DefaultConfigurationResolver struct {
	configurations []Configuration
}

// NewDefaultConfigurationResolver creates a new DefaultConfigurationResolver.
func NewDefaultConfigurationResolver(configurations []Configuration) ConfigurationResolver {
	return &DefaultConfigurationResolver{
		configurations: configurations,
	}
}

func (r *DefaultConfigurationResolver) ResolveConfiguration(ctx context.Context, chainName string) (Configuration, bool, error) {
	for _, configuration := range r.configurations {
		if configuration.ChainName == chainName {
			return configuration, true, nil
		}
	}

	return Configuration{}, false, nil
}
