package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kungze/wovenet/internal/logger"
	"github.com/kungze/wovenet/internal/restfulapi"
	"github.com/kungze/wovenet/internal/site"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Startup a wovenet site",
	Long: `This command will start a wovenet site. It will try to connect to other
wovenet sites and wait for other wovenet sites to connect to it. And then
create local sockets for remote applications, you can access these remote
applications by these local sockets.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := initConfig(); err != nil {
			return fmt.Errorf("init config error: %w", err)
		}
		if err := logger.InitLogging(); err != nil {
			return fmt.Errorf("init logging error: %w", err)
		}
		log := logger.GetDefault()
		var config site.Config
		err := viper.Unmarshal(&config)
		if err != nil {
			log.Error("failed to unmarshal the config into a struct", "error", err)
			return err
		}
		err = site.CheckAndSetDefaultConfig(&config)
		if err != nil {
			return fmt.Errorf("config error: %w", err)
		}
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		// Create a new site
		site, err := site.NewSite(ctx, config)
		if err != nil {
			log.Error("failed to crate site", "error", err)
			return err
		}
		if err := site.Start(ctx); err != nil {
			log.Error("failed to start site", "error", err)
			return err
		}
		log.Info("local site started")
		var apiConfig restfulapi.Config
		err = viper.UnmarshalKey("restfulApi", &apiConfig)
		if err != nil {
			log.Error("failed to unmarshal the config into a struct for restful api", "error", err)
			return err
		}
		if apiConfig.Enabled {
			log.Info("restful api server is enabled", "address", apiConfig.ListenAddr)
			go func() {
				defer cancel()
				err := restfulapi.SetupHTTPServer(&apiConfig, site, site.GetAppManager())
				if err != nil {
					log.Error("failed to setup restful api server", "error", err)
					return
				}
			}()
		}
		<-ctx.Done()
		log.Error("local site exit")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.wovenet/config.yaml)")
}

func initConfig() error {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".wovenet/config")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("fatal error config file: %s", err)
	}
	return nil
}
