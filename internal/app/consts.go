package app

const (
	SINGLE string = "single"
	RANGE  string = "range"
)

const (
	TCP  string = "tcp"
	UNIX string = "unix"
)

type AppType string

const (
	localAppType  AppType = "localExposedApp"
	remoteAppType AppType = "remoteApp"
)
