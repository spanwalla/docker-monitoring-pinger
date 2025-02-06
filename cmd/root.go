package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spanwalla/docker-monitoring-pinger/config"
	"github.com/spanwalla/docker-monitoring-pinger/internal/auth"
	"github.com/spanwalla/docker-monitoring-pinger/internal/collector"
	"github.com/spanwalla/docker-monitoring-pinger/internal/scheduler"
	"github.com/spanwalla/docker-monitoring-pinger/internal/sender"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string
var cfg *config.Config

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "docker-pinger",
	Short: "Periodically ping docker monitoring tool",
	Long:  `Periodically checks the status of docker containers running on the device and sends reports to the server.`,
	Run: func(cmd *cobra.Command, args []string) {
		setLogrus(cfg.Log.Level)
		name, password, err := cfg.Auth.GetCredentials()
		if err != nil {
			log.Fatalf("Error getting credentials: %v", err)
		}

		authClient := auth.NewClient(cfg.ApiUrl+"/api/v1/pingers", name, password)
		restSender := sender.NewRestSender(cfg.ApiUrl, authClient)
		dockerCollector, err := collector.NewDockerCollector()
		if err != nil {
			log.Fatalf("Error creating docker collector: %v", err)
		}

		taskSend := scheduler.NewTaskScheduler()
		taskSend.RunWithGracefulShutdown(dockerCollector, restSender, cfg.CronSpec)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.docker-pinger.yaml)")
}

func setLogrus(level string) {
	logrusLevel, err := log.ParseLevel(level)
	if err != nil {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(logrusLevel)
	}

	log.SetFormatter(&log.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
	})
}

func initConfig() {
	var err error
	cfg, err = config.LoadConfig(cfgFile)
	if err != nil {
		log.Fatalf("initConfig - config.LoadConfig: %v", err)
	}
}
