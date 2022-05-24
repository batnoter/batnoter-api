package cmd

import (
	"net"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/batnoter/batnoter-api/internal/config"
	"github.com/iamolegga/enviper"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "batnoter-api",
	Short: "A simple note taking app",
	Long:  `A simple web application to store and retrieve notes`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

var conf config.Config

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// here you will define your flags and configuration settings.
	// cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	initLogger()
	initConfig()

	// cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func initLogger() {
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}

func initConfig() {
	// enviper is a wrapper over viper that loads the config from file & overrides
	// them with env variables if available. If the config file is missing it simply ignores it
	e := enviper.New(viper.New())
	e.SetDefault("httpserver.host", "0.0.0.0")
	e.SetDefault("httpserver.port", "8080")

	var cfgFile string
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .config.yaml)")
	if cfgFile != "" {
		e.SetConfigFile(cfgFile)
	} else {
		e.AddConfigPath(".")
		e.SetConfigName(".config")
	}

	if err := e.Unmarshal(&conf); err == nil {
		logrus.Infof("using the config file: %s", e.ConfigFileUsed())
	}

	if conf.Database.URL != "" {
		// if url is set then the values of host, port, dbname, username, password, driver-name
		// will be overridden with their respective values from url string.
		logrus.Info("using db connection url")
		overrideDatabaseConfigs()
	}

	// evaluate configs with values set as environment variable
	if strings.HasPrefix(conf.HTTPServer.Host, "$") {
		logrus.Infof("evaluating http host config from env variable: %s", conf.HTTPServer.Host)
		conf.HTTPServer.Host = os.Expand(conf.HTTPServer.Host, os.Getenv)
	}
	if strings.HasPrefix(conf.HTTPServer.Port, "$") {
		logrus.Infof("evaluating http port config from env variable: %s", conf.HTTPServer.Port)
		conf.HTTPServer.Port = os.Expand(conf.HTTPServer.Port, os.Getenv)
	}
}

func overrideDatabaseConfigs() {
	u := parseDatabaseURL(conf.Database.URL)
	host, port, err := net.SplitHostPort(u.Host)
	if err != nil {
		logrus.Fatal("error occurred while retrieving db host and post: ", err)
	}
	conf.Database.Host = host
	conf.Database.Port = port
	conf.Database.DriverName = u.Scheme
	conf.Database.Username = u.User.Username()
	conf.Database.Password = password(u)
	conf.Database.DBName = path.Base(u.Path)
	if len(u.Query()["sslmode"]) > 0 {
		conf.Database.SSLMode = u.Query()["sslmode"][0]
	}
}

func parseDatabaseURL(dbURL string) url.URL {
	u, err := url.Parse(dbURL)
	if err != nil {
		logrus.Fatal("error occurred while parsing database url: ", err)
	}
	return *u
}

func password(u url.URL) string {
	password, isSet := u.User.Password()
	if !isSet {
		logrus.Fatal("database connection password must be set")
	}
	return password
}
