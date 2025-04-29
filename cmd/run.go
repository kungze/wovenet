package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/kungze/wovenet/internal/logger"
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
		ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer cancel()
		// Create a new site
		site, err := site.NewSite(ctx)
		if err != nil {
			log.Error("failed to crate site", "error", err)
			return err
		}
		if err := site.Start(ctx); err != nil {
			log.Error("failed to start site", "error", err)
			return err
		}
		log.Info("local site started")
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
