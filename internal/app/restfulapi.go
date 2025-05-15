package app

type LocalExposedAppModel struct {
	LocalExposedAppConfig
}

func (app *AppManager) GetLocalExposedApps() []LocalExposedAppModel {
	apps := []LocalExposedAppModel{}
	for _, app := range app.localExposedApps {
		apps = append(apps, LocalExposedAppModel{LocalExposedAppConfig: app.config})
	}
	return apps
}

func (app *AppManager) ShowLocalExposedApp(appName string) *LocalExposedAppModel {
	localApp, ok := app.localExposedApps[appName]
	if !ok {
		return nil
	}
	return &LocalExposedAppModel{LocalExposedAppConfig: localApp.config}
}

type RemoteAppModel struct {
	RemoteAppConfig
}

func (app *AppManager) GetRemoteApps() []RemoteAppModel {
	apps := []RemoteAppModel{}
	for _, app := range app.remoteApp {
		apps = append(apps, RemoteAppModel{RemoteAppConfig: app.config})
	}
	return apps
}

func (app *AppManager) ShowRemoteApp(appName string) *RemoteAppModel {
	for _, remoteApp := range app.remoteApp {
		if remoteApp.config.AppName == appName {
			return &RemoteAppModel{RemoteAppConfig: remoteApp.config}
		}
	}
	return nil
}
