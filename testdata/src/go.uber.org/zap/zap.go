package zap

type Logger struct{}
type SugaredLogger struct{}

func NewNop() *Logger {
	return &Logger{}
}

func (l *Logger) Sugar() *SugaredLogger {
	return &SugaredLogger{}
}

func (l *Logger) Debug(_ string, _ ...any)  {}
func (l *Logger) Info(_ string, _ ...any)   {}
func (l *Logger) Warn(_ string, _ ...any)   {}
func (l *Logger) Error(_ string, _ ...any)  {}
func (l *Logger) DPanic(_ string, _ ...any) {}
func (l *Logger) Panic(_ string, _ ...any)  {}
func (l *Logger) Fatal(_ string, _ ...any)  {}

func (s *SugaredLogger) Debugf(_ string, _ ...any)  {}
func (s *SugaredLogger) Infof(_ string, _ ...any)   {}
func (s *SugaredLogger) Warnf(_ string, _ ...any)   {}
func (s *SugaredLogger) Errorf(_ string, _ ...any)  {}
func (s *SugaredLogger) DPanicf(_ string, _ ...any) {}
func (s *SugaredLogger) Panicf(_ string, _ ...any)  {}
func (s *SugaredLogger) Fatalf(_ string, _ ...any)  {}

func (s *SugaredLogger) Debugw(_ string, _ ...any)  {}
func (s *SugaredLogger) Infow(_ string, _ ...any)   {}
func (s *SugaredLogger) Warnw(_ string, _ ...any)   {}
func (s *SugaredLogger) Errorw(_ string, _ ...any)  {}
func (s *SugaredLogger) DPanicw(_ string, _ ...any) {}
func (s *SugaredLogger) Panicw(_ string, _ ...any)  {}
func (s *SugaredLogger) Fatalw(_ string, _ ...any)  {}
