package common

import "go.uber.org/zap"

var Logger *zap.Logger

func InitializeLogger() {
	if IsProd() {
		initializeProdLogger()
	} else {
		initializeDevLogger()
	}
}

func initializeProdLogger() {

}

func initializeDevLogger() {
	Logger = zap.Must(zap.NewDevelopment())
}
