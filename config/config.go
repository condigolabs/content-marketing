package config

import (
	"github.com/sirupsen/logrus"
	"os"
	"strconv"
)

var (
	LogLevel  = GetStrEnv("LOGLEVEL", "info")
	LogFormat = GetStrEnv("LOGFORMAT", "json")
)

func LogInit() {
	level, _ := logrus.ParseLevel(LogLevel)
	logrus.SetLevel(level)
	logrus.SetOutput(os.Stdout)
	if LogFormat == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{})
	}
}
func GetStrEnv(name, defaultVal string) string {
	res := os.Getenv(name)
	if len(res) == 0 {
		return defaultVal
	}
	return res
}

func GetInt64(name string, defaultVal int64) int64 {
	res := os.Getenv(name)
	if len(res) == 0 {
		return defaultVal
	}
	n, err := strconv.ParseInt(res, 10, 64)
	if err != nil {
		return defaultVal
	}
	return n
}
