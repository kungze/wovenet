package restfulapi

import (
	"crypto/tls"
	"io"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/kungze/wovenet/internal/app"
	"github.com/kungze/wovenet/internal/logger"
	_ "github.com/kungze/wovenet/internal/restfulapi/docs"
	"github.com/kungze/wovenet/internal/site"

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

//	@title			Wovenet API
//	@version		1.0
//	@description	This is wovenet api definitions
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	JffYang
//	@contact.url	https://github.com/kungze/wovenet/issues
//	@contact.email	jeffyang512@163.com

//	@license.name	MIT
//	@license.url	https://github.com/kungze/wovenet/blob/main/LICENSE

//	@BasePath	/api/v1

//	@securityDefinitions.basic	BasicAuth

func SetupRouter(config *Config, site *site.Site, app *app.AppManager) (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)
	var output io.Writer
	var err error
	if config.Logger.File != "" {
		output, err = os.OpenFile(config.Logger.File, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
		if err != nil {
			return nil, err
		}
	} else {
		output = os.Stdout
	}
	logger := gin.LoggerWithWriter(output)
	r := gin.New()
	r.Use(logger, gin.Recovery())

	// Ping test
	r.GET("/ping", func(c *gin.Context) {
		c.String(http.StatusOK, "pong")
	})

	// Login
	r.POST("/login", func(c *gin.Context) {
		var loginRequest struct {
			Username string `json:"username" binding:"required"`
			Password string `json:"password" binding:"required"`
		}
		if err := c.ShouldBindJSON(&loginRequest); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		for _, auth := range config.Auth.BasicAuth {
			if loginRequest.Username == auth.User && loginRequest.Password == auth.Password {
				c.JSON(http.StatusOK, gin.H{"message": "Login successful"})
				return
			}
		}
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
	})

	handlers := []gin.HandlerFunc{}
	if len(config.Auth.BasicAuth) > 0 {
		account := gin.Accounts{}
		for _, auth := range config.Auth.BasicAuth {
			account[auth.User] = auth.Password
		}
		handlers = append(handlers, gin.BasicAuth(account))
	}

	authorized := r.Group("/api/v1", handlers...)
	{
		c := newController(site, app)
		// remote sites
		authorized.GET("/remoteSites", c.ListRemoteSites)
		authorized.GET("/remoteSites/:siteName", c.ShowRemoteSite)
		// local exposed apps
		authorized.GET("/localExposedApps", c.ListLocalExposedApps)
		authorized.GET("/localExposedApps/:appName", c.ShowLocalExposedApp)
		// remote apps
		authorized.GET("/remoteApps", c.ListRemoteApps)
		authorized.GET("/remoteApps/:appName", c.ShowRemoteApp)
	}

	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r, nil
}

// Setup http server
func SetupHTTPServer(config *Config, site *site.Site, app *app.AppManager) error {
	log := logger.GetDefault()
	r, err := SetupRouter(config, site, app)
	if err != nil {
		log.Error("failed to setup router", "error", err)
		return err
	}
	server := &http.Server{
		Addr:    config.ListenAddr,
		Handler: r,
	}

	if config.Tls.Enabled {
		cert, err := tls.LoadX509KeyPair(config.Tls.Cert, config.Tls.Key)
		if err != nil {
			log.Error("failed to load tls certificate", "error", err)
			return err
		}
		server.TLSConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
		return server.ListenAndServeTLS("", "")
	}

	return server.ListenAndServe()
}
