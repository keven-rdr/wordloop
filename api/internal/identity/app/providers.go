package app

import "sort"

const DefaultProvider = "oidc"

type Providers map[string]Provider

func resolveProvider(single Provider, named Providers, name string) (Provider, string, bool) {
	if name == "" {
		name = DefaultProvider
	}
	if p, ok := named[name]; ok {
		return p, name, true
	}
	if name == DefaultProvider && single != nil {
		return single, name, true
	}
	return nil, name, false
}

func providerNames(single Provider, named Providers) []string {
	names := make([]string, 0, len(named)+1)
	for name := range named {
		names = append(names, name)
	}
	if _, ok := named[DefaultProvider]; !ok && single != nil {
		names = append(names, DefaultProvider)
	}
	sort.Strings(names)
	return names
}
