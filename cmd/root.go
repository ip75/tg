package cmd

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	logFile string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "tg",
	Short: "Manage and serve telegram groups, channels, and bots.",
	Long: `Populate topics in group and serve new media in telegram.
New media will appear in telegram group topic when tag was added to media metadata.
Tag linked directly to telegram topic to publish a message in this topic.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
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

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.tg.yaml)")
	rootCmd.PersistentFlags().StringVarP(&logFile, "log", "l", "tg.log", "log file (default is tg.log)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".tg" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".tg")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println(fmt.Errorf("read config: %w", err))
	}
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	initLogger()
}

func initLogger() {
	fileLogger, err := os.OpenFile(
		logFile,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	if err != nil {
		log.Panic().Err(err).Str("file", logFile).Msg("open log file")
	}
	writers := zerolog.MultiLevelWriter(zerolog.ConsoleWriter{Out: os.Stderr}, fileLogger)

	loglevel := zerolog.DebugLevel
	switch viper.GetString("server.loglevel") {
	case "debug":
		loglevel = zerolog.DebugLevel
	case "info":
		loglevel = zerolog.InfoLevel
	case "error":
		loglevel = zerolog.ErrorLevel
	}

	log.Logger = zerolog.New(writers).Level(loglevel).With().Timestamp().Logger()
}
