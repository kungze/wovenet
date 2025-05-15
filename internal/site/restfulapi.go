package site

import (
	"github.com/kungze/wovenet/internal/app"
	"github.com/kungze/wovenet/internal/tunnel"
)

type RemoteSiteModel struct {
	SiteName      string                `json:"siteName"`
	TunnelSockets []tunnel.SocketInfo   `json:"tunnelSockets"`
	ExposedApps   []app.LocalExposedApp `json:"exposedApps"`
}

func (s *Site) GetRemoteSites() []RemoteSiteModel {
	sites := []RemoteSiteModel{}
	s.remoteSites.Range(func(key, value any) bool {
		siteName := key.(string)
		siteInfo := value.(*siteInfo)
		sites = append(sites, RemoteSiteModel{SiteName: siteName, TunnelSockets: siteInfo.TunnelListenerSockets, ExposedApps: siteInfo.ExposedApps})
		return true
	})
	return sites
}

func (s *Site) ShowRemoteSite(siteName string) *RemoteSiteModel {
	value, ok := s.remoteSites.Load(siteName)
	if !ok {
		return nil
	}
	return &RemoteSiteModel{
		SiteName:      siteName,
		TunnelSockets: value.(*siteInfo).TunnelListenerSockets,
		ExposedApps:   value.(*siteInfo).ExposedApps,
	}
}

func (s *Site) GetAppManager() *app.AppManager {
	return s.appManager
}
