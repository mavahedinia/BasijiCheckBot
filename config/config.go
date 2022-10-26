package configs

import (
	"os"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type ConfigChangeListener interface {
	OnConfigChanged()
}

var (
	changeListeners []ConfigChangeListener
	changeMutex     = &sync.Mutex{}
)

func NewConfig() *viper.Viper {
	var configDir string
	configFileName := "config-dev"

	envConfName, ok := os.LookupEnv("BOT_CONFIG_FILE")
	if ok {
		configFileName = envConfName
	}

	configDir, ok = os.LookupEnv("BOT_CONFIG_DIR")
	if !ok {
		configDir = "./config/"
	}

	config := viper.New()

	if err := config.BindEnv("telegram.bot.token", "TELEGRAM_BOT_TOKEN"); err != nil {
		logrus.WithError(err).Fatal("Failed to bind TELEGRAM_BOT_TOKEN")
	}

	config.AddConfigPath(configDir)
	config.SetConfigName(configFileName)
	if err := config.ReadInConfig(); err != nil {
		logrus.WithError(err).Fatal("Configuration Error")
		return nil
	}

	config.WatchConfig()
	config.OnConfigChange(configChanged)

	logrus.Info("Configuration is loaded!")
	AdjustLogLevel(config.GetString("log.level"))
	return config
}

func AdjustLogLevel(logLevelStr string) {
	logLevel, err := logrus.ParseLevel(logLevelStr)
	logrus.SetReportCaller(true)
	if err != nil {
		logrus.WithError(err).WithField("logLevel", logLevelStr).Error("failed to AdjustLogLevel")
		logrus.SetLevel(logrus.DebugLevel)
		return
	}
	logrus.WithField("LogLevel", logLevelStr).Info("logLevel has been set")
	logrus.SetLevel(logLevel)
}

func AddToChangeListener(listener ConfigChangeListener) {
	changeListeners = append(changeListeners, listener)
}

func configChanged(_ fsnotify.Event) {
	changeMutex.Lock()
	for _, listener := range changeListeners {
		listener.OnConfigChanged()
	}
	changeMutex.Unlock()
}
