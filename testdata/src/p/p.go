package p

import (
	"fmt"
	"log/slog"

	"go.uber.org/zap"
)

func slogCases(password string, token string) {
	slog.Info("Starting server on port 8080")      // want "log message should start with a lowercase letter"
	slog.Error("ошибка подключения к базе данных") // want "log message should be in English only"
	slog.Warn("warning: something went wrong...")  // want "log message should not contain special characters or emoji"
	slog.Info("api_key=" + password)               // want "log message may expose sensitive data"
	slog.Info("token: " + token)                   // want "log message may expose sensitive data"
	slog.Info("server started")                    // ok
	slog.Info("password: hidden")                  // want "log message should not contain special characters or emoji"
	slog.Info("token=123")                         // want "log message should not contain special characters or emoji"
	slog.Info("пароль получен 🚀")                  // want "log message should be in English only" "log message should not contain special characters or emoji"
	slog.InfoContext(nil, "Connection failed!!!")  // want "log message should start with a lowercase letter" "log message should not contain special characters or emoji"
	slog.Default().Warn("Something went wrong")    // want "log message should start with a lowercase letter"
	slog.Info(fmt.Sprintf("token=%s", token))      // want "log message may expose sensitive data"
}

func zapCases(apiKey string, token string) {
	logger := zap.NewNop()
	sugar := logger.Sugar()

	logger.Info("Starting server")                             // want "log message should start with a lowercase letter"
	logger.Error("connection failed!!!")                       // want "log message should not contain special characters or emoji"
	logger.Warn("запуск сервера")                              // want "log message should be in English only"
	logger.Debug("api_key=" + apiKey)                          // want "log message may expose sensitive data"
	logger.Info("server started")                              // ok
	sugar.Infof("Failed to connect: %s", token)                // want "log message should start with a lowercase letter" "log message should not contain special characters or emoji"
	sugar.Infof("failed to connect: %s", token)                // want "log message should not contain special characters or emoji"
	sugar.Infof("ошибка: %s", token)                           // want "log message should be in English only" "log message should not contain special characters or emoji"
	sugar.Infow("token: "+token, "request_id", "abc")          // want "log message may expose sensitive data"
	sugar.Debugw("api request completed", "request_id", "123") // ok
}

func customSensitiveCases(merchantPin string, otpCode string) {
	slog.Info("merchant_pin: " + merchantPin) // want "log message should not contain special characters or emoji"
	slog.Info("otp code: " + otpCode)         // want "log message should not contain special characters or emoji"
}
