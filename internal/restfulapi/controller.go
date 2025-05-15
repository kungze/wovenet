package restfulapi

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kungze/wovenet/internal/app"
	"github.com/kungze/wovenet/internal/site"
)

type controller struct {
	site *site.Site
	app  *app.AppManager
}

// ShowRemoteSite godoc
//
//	@Summary		Get a remote site's details
//	@Description	Get a remote site's details by siteName
//	@Tags			remoteSites
//	@Accept			json
//	@Produce		json
//	@Param			siteName	path		string	true	"remote site name"
//	@Success		200	{object}	site.RemoteSiteModel
//	@Failure		400	{object}	HTTPError
//	@Failure		401	{object}	HTTPError
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/remoteSites/{siteName} [get]
func (c *controller) ShowRemoteSite(ctx *gin.Context) {
	name := ctx.Param("siteName")
	remoteSite := c.site.ShowRemoteSite(name)
	if remoteSite == nil {
		NewError(ctx, http.StatusNotFound, fmt.Errorf("remote site: %s not found", name))
		return
	}
	ctx.JSON(http.StatusOK, remoteSite)
}

// ListRemoteSites godoc
//
//	@Summary		List remote sites
//	@Description	Get all remote sites
//	@Tags			remoteSites
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		site.RemoteSiteModel
//	@Failure		401	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/remoteSites [get]
func (c *controller) ListRemoteSites(ctx *gin.Context) {
	remoteSites := c.site.GetRemoteSites()
	ctx.JSON(http.StatusOK, remoteSites)
}

// ShowLocalExposedApp godoc
//
//	@Summary		Show a local exposed app
//	@Description	Show a local exposed app by appName
//	@Tags			localExposedApps
//	@Accept			json
//	@Produce		json
//	@Param			appName	path		string	true	"local exposed app name"
//	@Success		200	{object}	app.LocalExposedAppModel
//	@Failure		400	{object}	HTTPError
//	@Failure		401	{object}	HTTPError
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/localExposedApps/{appName} [get]
func (c *controller) ShowLocalExposedApp(ctx *gin.Context) {
	appName := ctx.Param("appName")
	app := c.app.ShowLocalExposedApp(appName)
	if app == nil {
		NewError(ctx, http.StatusNotFound, fmt.Errorf("local exposed app: %s not found", appName))
		return
	}
	ctx.JSON(http.StatusOK, app)
}

// ListLocalExposedApps godoc
//
//	@Summary		List local exposed apps
//	@Description	List all local exposed apps
//	@Tags			localExposedApps
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		app.LocalExposedAppModel
//	@Failure		401	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/localExposedApps [get]
func (c *controller) ListLocalExposedApps(ctx *gin.Context) {
	apps := c.app.GetLocalExposedApps()
	ctx.JSON(http.StatusOK, apps)
}

// ShowRemoteApp godoc
//
//	@Summary		Show a remote app
//	@Description	Show a remote app by appName
//	@Tags			remoteApps
//	@Accept			json
//	@Produce		json
//	@Param			appName	path		string	true	"remote app name"
//	@Success		200	{object}	app.RemoteAppModel
//	@Failure		400	{object}	HTTPError
//	@Failure		401	{object}	HTTPError
//	@Failure		404	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/remoteApps/{appName} [get]
func (c *controller) ShowRemoteApp(ctx *gin.Context) {
	appName := ctx.Param("appName")
	app := c.app.ShowRemoteApp(appName)
	if app == nil {
		NewError(ctx, http.StatusNotFound, fmt.Errorf("remote app: %s not found", appName))
		return
	}
	ctx.JSON(http.StatusOK, app)
}

// ListRemoteApps godoc
//
//	@Summary		List remote apps
//	@Description	List all remote apps
//	@Tags			remoteApps
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}		app.RemoteAppModel
//	@Failure		401	{object}	HTTPError
//	@Failure		500	{object}	HTTPError
//	@Router			/remoteApps [get]
func (c *controller) ListRemoteApps(ctx *gin.Context) {
	apps := c.app.GetRemoteApps()
	ctx.JSON(http.StatusOK, apps)
}

func newController(site *site.Site, app *app.AppManager) *controller {
	return &controller{site: site, app: app}
}
