package config

// Config is keepsake.yaml
type Config struct {
	Repository string `json:"repository"`

	Storage string `json:"storage"` // deprecated
}

func getDefaultConfig(workingDir string) *Config {
	// should match defaults in config.py
	return &Config{}
}
