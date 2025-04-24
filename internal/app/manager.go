package app

import (
	"context"
	"fmt"
	"io"

	"gihtub.com/kungze/wovenet/internal/logger"
)

type AppManager struct {
	localExposedApps map[string]*localApp
	remoteApp        []*remoteApp
}

func (am *AppManager) GetExposedApps() []LocalExposedApp {
	apps := []LocalExposedApp{}
	for _, app := range am.localExposedApps {
		apps = append(apps, LocalExposedApp{Id: app.config.Id})
	}
	return apps
}

// ConnectToLocalApp get a connection which connect to the local app
func (am *AppManager) ConnectToLocalApp(appId string) (io.ReadWriteCloser, error) {
	log := logger.GetDefault()
	app, ok := am.localExposedApps[appId]
	if !ok {
		log.Error("local app can not found", "appId", appId)
		return nil, fmt.Errorf("app: %s can not found", appId)
	}
	return app.GetConnection()
}

// ProcessNewRemoteSite when a new remote site connected successfully, we
// need to start the listeners for local sockets which for remote apps that
// located in this new remote site
func (am *AppManager) ProcessNewRemoteSite(ctx context.Context, remoteSite string, exposedApps []LocalExposedApp, callback ClientConnectedCallback) {
	log := logger.GetDefault()
	for _, remoteApp := range am.remoteApp {
		if remoteApp.Active() {
			continue
		}
		for _, eApp := range exposedApps {
			if remoteApp.config.SiteName == remoteSite && remoteApp.config.RemoteAppId == eApp.Id {
				if err := remoteApp.listen(ctx, callback); err != nil {
					log.Error("failed to start local socket listener for remote app", "localSocket", remoteApp.config.LocalSocket, "remoteSite", remoteSite, "remoteAppId", remoteApp.config.RemoteAppId, "error", err)
					continue
				}
			}
		}
	}
}

// ProcessRemoteSiteGone when a remote site is disconnected, we need to
// stop the listeners which related to the remote apps
func (am *AppManager) ProcessRemoteSiteGone(remoteSite string) {
	for _, remoteApp := range am.remoteApp {
		if remoteApp.config.SiteName == remoteSite && remoteApp.Active() {
			remoteApp.stop()
		}
	}
}

func NewAppManager(ctx context.Context, localexposedApps []*LocalExposedAppConfig, remoteApps []*RemoteAppConfig) (*AppManager, error) {
	am := AppManager{localExposedApps: map[string]*localApp{}}
	for _, exposedApp := range localexposedApps {
		a := newLocalApp(*exposedApp)
		am.localExposedApps[exposedApp.Id] = a
	}

	for _, remoteApp := range remoteApps {
		a := newRemoteApp(*remoteApp)
		am.remoteApp = append(am.remoteApp, a)
	}

	return &am, nil
}
