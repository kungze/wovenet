package app

import (
	"context"
	"fmt"

	"github.com/kungze/wovenet/internal/logger"
	"github.com/kungze/wovenet/internal/tunnel"
)

type AppManager struct {
	localExposedApps map[string]*localApp
	remoteApp        []*remoteApp
}

func (am *AppManager) GetExposedApps() []LocalExposedApp {
	apps := []LocalExposedApp{}
	for _, app := range am.localExposedApps {
		apps = append(apps, LocalExposedApp{Name: app.config.AppName})
	}
	return apps
}

// TransferDataToLocalApp transfer data between tunnel stream and local app service
// appName the local app's name which the remote site's app client want to connect to
// socket if the local app's mode is range (usually have multiple socket), the socket specifies that connect to which socket
// stream tunnel stream
// remainingData the extra data except handshake data during handshake period, it come from remote site's app client, we need to
// write it to local app service
func (am *AppManager) TransferDataToLocalApp(appName string, socket string, stream tunnel.Stream, remainingData []byte) error {
	log := logger.GetDefault()
	app, ok := am.localExposedApps[appName]
	if !ok {
		log.Error("local app can not found", "localApp", appName)
		return fmt.Errorf("app: %s can not found", appName)
	}
	if err := app.StartDataConverter(stream, socket, remainingData); err != nil {
		log.Error("failed to start data converter", "error", err, "localApp", appName)
	}
	return nil
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
		for _, exposedApp := range exposedApps {
			if remoteApp.config.SiteName == remoteSite && remoteApp.config.AppName == exposedApp.Name {
				if err := remoteApp.listen(ctx, callback); err != nil {
					log.Error("failed to start local socket listener for remote app", "localSocket", remoteApp.config.LocalSocket, "remoteSite", remoteSite, "appName", remoteApp.config.AppName, "error", err)
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

func NewAppManager(ctx context.Context, localExposedApps []*LocalExposedAppConfig, remoteApps []*RemoteAppConfig) (*AppManager, error) {
	am := AppManager{localExposedApps: map[string]*localApp{}}
	for _, exposedApp := range localExposedApps {
		a := newLocalApp(*exposedApp)
		am.localExposedApps[exposedApp.AppName] = a
	}

	for _, remoteApp := range remoteApps {
		a := newRemoteApp(*remoteApp)
		am.remoteApp = append(am.remoteApp, a)
	}

	return &am, nil
}
