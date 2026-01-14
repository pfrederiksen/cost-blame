package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

var (
	cfgFile string
	debug   bool
	logger  *zap.Logger
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "cost-blame",
	Short: "Attribute AWS cost spikes to services, tags, and resources",
	Long: `cost-blame is a local-first CLI tool that helps CloudOps teams
attribute AWS spend increases across accounts, regions, and tags.

Use AWS Cost Explorer to detect spikes, identify new spenders,
and drill down to specific resources causing cost changes.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		initLogger()
	},
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.cost-blame.yaml)")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug logging")
	rootCmd.PersistentFlags().String("profile", "", "AWS profile")
	rootCmd.PersistentFlags().String("region", "us-east-1", "AWS region")

	viper.BindPFlag("profile", rootCmd.PersistentFlags().Lookup("profile"))
	viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region"))
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			viper.AddConfigPath(home)
			viper.SetConfigType("yaml")
			viper.SetConfigName(".cost-blame")
		}
	}

	viper.AutomaticEnv()
	viper.ReadInConfig()
}

func initLogger() {
	var err error
	if debug {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
}

func getLogger() *zap.Logger {
	if logger == nil {
		initLogger()
	}
	return logger
}
