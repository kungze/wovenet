package app

type LocalExposedAppConfig struct {
	Id     string `mapstructure:"id"`
	Socket string `mapstructure:"socket"`
}

type RemoteAppConfig struct {
	RemoteAppId string `mapstructure:"remoteAppId"`
	LocalSocket string `mapstructure:"localSocket"`
	SiteName    string `mapstructure:"siteName"`
}
