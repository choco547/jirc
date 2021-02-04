package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

const (
	defaultJiraURL = "http://jira.yourcoman.com"
)

var (
	cfgFile      string
	jiraEndPoint string
	jiraUser     string
	jiraPass     string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "jirc",
	Short: "Deployment management with Jira",
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	pflags := rootCmd.PersistentFlags()

	pflags.StringVar(&cfgFile, "config", "", "config file (default is $HOME/.jirc.yaml)")

	pflags.StringVarP(&jiraEndPoint, "jira", "j", defaultJiraURL, "Jira base URL")
	pflags.StringVarP(&jiraUser, "user", "u", "", "Jira username")
	pflags.StringVarP(&jiraPass, "pass", "p", "", "Jira password")

	viper.BindPFlag("jira.url", pflags.Lookup("jira"))
	viper.BindPFlag("jira.user", pflags.Lookup("user"))
	viper.BindPFlag("jira.pass", pflags.Lookup("pass"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// readGlobalConfig()

	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".jirc" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".jirc")
	}

	viper.SetEnvPrefix("JIRC")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.MergeInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}

func readGlobalConfig() {
	configFile := filepath.Join(string(filepath.Separator), "etc", "jirc", "config.yml")
	finfo, err := os.Stat(configFile)

	if err != nil {
		return
	}

	if !finfo.Mode().IsRegular() {
		return
	}

	viper.SetConfigFile(configFile)
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	} else {
		fmt.Println(err)
	}
}
