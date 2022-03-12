package applicationconfig

import "github.com/vivekweb2013/batnoter/internal/config"

type ApplicationConfig struct {
	Config config.Config
}

func NewApplicationConfig(config config.Config) *ApplicationConfig {
	return &ApplicationConfig{
		Config: config,
	}
}
