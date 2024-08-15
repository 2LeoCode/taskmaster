package config

type ConfigLoader func() (*Config, error)

func NewConfigLoader(path string) ConfigLoader {
	return func() (*Config, error) {
		return parse(path)
	}
}
