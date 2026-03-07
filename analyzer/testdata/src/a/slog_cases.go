package a

import (
	"log/slog"
)

func slogCases(password string, token string) {
	slog.Info("Starting server on port 8080")       // want "log message should start with a lowercase letter"
	slog.Error("ошибка подключения к базе данных")  // want "log message should be in English only"
	slog.Warn("warning: something went wrong...")   // want "log message should not contain special characters or emoji"
	slog.Info("api_key=" + password)                // want "log message may expose sensitive data"
	slog.Info("token: " + token)                    // want "log message may expose sensitive data"
	slog.Info("failed to connect to database")      // ok
	slog.Info("token validated")                    // ok
	slog.Info("server started")                     // ok
	slog.Info("password: hidden")                   // want "log message may expose sensitive data"
	slog.Info("server started 🚀")                   // want "log message should not contain special characters or emoji"
	slog.Info("запуск сервера")                     // want "log message should be in English only"
	slog.Info("Server started")                     // want "log message should start with a lowercase letter"
	slog.Info("normal english message")             // ok
	slog.InfoContext(nil, "Connection failed!!!")   // want "log message should start with a lowercase letter" "log message should not contain special characters or emoji"
	slog.InfoContext(nil, "connection failed")      // ok
	slog.InfoContext(nil, "jwt: "+token)            // want "log message may expose sensitive data"
	slog.Default().Error("Private_key=" + password) // want "log message should start with a lowercase letter" "log message may expose sensitive data"
	slog.Default().Info("private_key: abc")         // want "log message may expose sensitive data"
	slog.Default().Warn("something went wrong")     // ok
	slog.Default().Warn("something went wrong...")  // want "log message should not contain special characters or emoji"
	slog.Default().Warn("Something went wrong")     // want "log message should start with a lowercase letter"
}
