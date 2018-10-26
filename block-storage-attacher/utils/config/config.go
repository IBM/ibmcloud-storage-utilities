package config

import (
	"github.com/BurntSushi/toml"
	"go.uber.org/zap"
	"os"
	"strconv"
	"strings"
)

func getEnv(key string) string {
	return os.Getenv(strings.ToUpper(key))
}

// GetGoPath ...
func GetGoPath() string {
	if goPath := getEnv("GOPATH"); goPath != "" {
		return goPath
	}
	return ""
}

// ParseConfig ...
func ParseConfig(filePath string, conf interface{}, logger zap.Logger) {
	if _, err := toml.DecodeFile(filePath, conf); err != nil {
		logger.Fatal("error parsing config file", zap.Error(err))
	}
}

// GetConfigString ...
func GetConfigString(envKey, defaultConf string) string {
	if val := getEnv(envKey); val != "" {
		return val
	}
	return defaultConf
}

// GetConfigInt ...
func GetConfigInt(envKey string, defaulfConf int, logger zap.Logger) int {
	if val := getEnv(envKey); val != "" {
		if envInt, err := strconv.Atoi(val); err == nil {
			return envInt
		}
		logger.Error("error parsing env val to int", zap.String("env", envKey))
	}
	return defaulfConf
}

// GetConfigBool ...
func GetConfigBool(envKey string, defaultConf bool, logger zap.Logger) bool {
	if val := getEnv(envKey); val != "" {
		if envBool, err := strconv.ParseBool(val); err == nil {
			return envBool
		}
		logger.Error("error parsing env val to bool", zap.String("env", envKey))
	}
	return defaultConf
}

// GetConfigStringList ...
func GetConfigStringList(envKey string, defaultConf string, logger zap.Logger) []string {
	// Assume env var is a list of strings separated by ','
	val := defaultConf

	if getEnv(envKey) != "" {
		val = getEnv(envKey)
	}

	val = strings.Replace(val, " ", "", -1)
	return strings.Split(val, ",")
}

type Volume struct {
	Iqn      string `json:"iqn,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	Target   string `json:"target,omitempty"`
	Lunid    int    `json:"lunid,omitempty"`
	Nodeip   string `json:"nodeip,omitempty"`
}
