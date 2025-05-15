package restfulapi

type BasicAuth struct {
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}

type Auth struct {
	BasicAuth []BasicAuth `mapstructure:"basicAuth"`
}

type Logger struct {
	File string `mapstructure:"file"`
}

type TLS struct {
	Enabled bool   `mapstructure:"enabled"`
	Key     string `mapstructure:"key"`
	Cert    string `mapstructure:"cert"`
}

type Config struct {
	Enabled    bool   `mapstructure:"enabled"`
	ListenAddr string `mapstructure:"listenAddr"`
	Logger     Logger `mapstructure:"logger"`
	Tls        TLS    `mapstructure:"tls"`
	Auth       Auth   `mapstructure:"auth"`
}
