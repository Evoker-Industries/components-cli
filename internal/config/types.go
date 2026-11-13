package config

type ForgeAuth struct {
	Secret string `json:"secret,omitempty"`
}

type ForgeConfig struct {
	Type string     `json:"type"`
	URL  string     `json:"url"`
	API  string     `json:"api,omitempty"`
	Auth *ForgeAuth `json:"auth,omitempty"`
}

type ComponentRef struct {
	Ref string `json:"ref"`
}

type RegistryFile struct {
	Components       map[string]ComponentRef `json:"components,omitempty"`
	LegacyComponents map[string]ComponentRef `json:"Components,omitempty"`
	Forges           map[string]ForgeConfig  `json:"forges,omitempty"`
}

type Registry struct {
	Path       string
	Dir        string
	Components map[string]ComponentRef
	Forges     map[string]ForgeConfig
}
