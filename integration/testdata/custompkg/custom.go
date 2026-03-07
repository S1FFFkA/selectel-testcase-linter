package custompkg

import "log/slog"

func Run(merchantPin string, otpCode string) {
	slog.Info("merchant_pin: " + merchantPin)
	slog.Info("otp code: " + otpCode)
}
