package config

import (
	"github.com/charmbracelet/log"
)

func WriteLogLevel(level string) error {
	sessionConfig, err := GetSessionConfig()
	if err != nil {
		log.Errorf("Failed to get session config: %s", err.Error())
		return err
	}
	sessionConfig.LogLevel = level
	err = sessionConfig.Write()
	if err != nil {
		log.Errorf("Failed to write session config: %s", err.Error())
		return err
	}

	return nil
}

func ActivateLogLevel() {
	sessionConfig, err := GetSessionConfig()
	if err != nil {
		log.Errorf("Failed to get session config: %s", err.Error())
		return
	}
	logLevel, err := log.ParseLevel(sessionConfig.LogLevel)
	log.SetLevel(logLevel)
}
