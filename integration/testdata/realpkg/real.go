package realpkg

import (
	"log/slog"

	"go.uber.org/zap"
)

func Run(password string, token string) {
	logger := zap.NewNop()
	sugar := logger.Sugar()

	slog.Info("Starting server")
	slog.Error("ошибка подключения")
	slog.Warn("connection failed!!!")
	slog.Info("token: " + token)

	logger.Info("Starting api")
	logger.Warn("warning: unstable")
	logger.Error("api_key=" + password)

	sugar.Infof("Failed to connect: %s", token)
}
