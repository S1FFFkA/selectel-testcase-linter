package zap

type Logger struct{}

type SugaredLogger struct{}

type Field struct{}

func NewNop() *Logger { return &Logger{} }

func (l *Logger) Sugar() *SugaredLogger { return &SugaredLogger{} }

func (l *Logger) Debug(msg string, fields ...Field)  {}
func (l *Logger) Info(msg string, fields ...Field)   {}
func (l *Logger) Warn(msg string, fields ...Field)   {}
func (l *Logger) Error(msg string, fields ...Field)  {}
func (l *Logger) DPanic(msg string, fields ...Field) {}
func (l *Logger) Panic(msg string, fields ...Field)  {}
func (l *Logger) Fatal(msg string, fields ...Field)  {}

func (l *SugaredLogger) Debugf(template string, args ...interface{})  {}
func (l *SugaredLogger) Infof(template string, args ...interface{})   {}
func (l *SugaredLogger) Warnf(template string, args ...interface{})   {}
func (l *SugaredLogger) Errorf(template string, args ...interface{})  {}
func (l *SugaredLogger) DPanicf(template string, args ...interface{}) {}
func (l *SugaredLogger) Panicf(template string, args ...interface{})  {}
func (l *SugaredLogger) Fatalf(template string, args ...interface{})  {}

func (l *SugaredLogger) Debugw(msg string, keysAndValues ...interface{})  {}
func (l *SugaredLogger) Infow(msg string, keysAndValues ...interface{})   {}
func (l *SugaredLogger) Warnw(msg string, keysAndValues ...interface{})   {}
func (l *SugaredLogger) Errorw(msg string, keysAndValues ...interface{})  {}
func (l *SugaredLogger) DPanicw(msg string, keysAndValues ...interface{}) {}
func (l *SugaredLogger) Panicw(msg string, keysAndValues ...interface{})  {}
func (l *SugaredLogger) Fatalw(msg string, keysAndValues ...interface{})  {}
