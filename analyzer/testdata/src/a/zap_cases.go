package a

import "go.uber.org/zap"

func zapCases(apiKey string, token string) {
	logger := zap.NewNop()
	sugar := logger.Sugar()

	logger.Info("Starting server")       // want "log message should start with a lowercase letter"
	logger.Error("connection failed!!!") // want "log message should not contain special characters or emoji"
	logger.Warn("запуск сервера")        // want "log message should be in English only"
	logger.Debug("api_key=" + apiKey)    // want "log message may expose sensitive data"
	logger.Info("token validated")       // ok
	logger.Info("server started")        // ok
	logger.Info("password: hidden")      // want "log message may expose sensitive data"

	sugar.Infof("Failed to connect: %s", token)                // want "log message should start with a lowercase letter"
	sugar.Infof("failed to connect %s", token)                 // ok
	sugar.Infow("token: "+token, "request_id", "abc")          // want "log message may expose sensitive data"
	sugar.Debugw("api request completed", "request_id", "123") // ok
	sugar.Warnf("warning: bad input")                          // ok
}
