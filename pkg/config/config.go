package config

type Config struct {
	Debug      bool   `mapstructure:"debug"`
	CacheDir   string `mapstructure:"cache_dir"`
	Repository string `mapstructure:"repository"`
}
