package logging

import (
	"fmt"
	"log"
	"os"
)

type Logger struct {
	infoLogger *log.Logger
	warnLogger *log.Logger
	errLogger  *log.Logger
	debugLogger *log.Logger
	DebugMode bool
}

func NewLogger(debug bool) Logger {

	if debug {
		fmt.Printf("Logger initialized in debug mode.")
	}

	return Logger{
		infoLogger: log.New(os.Stdout, "INFO: ", log.LstdFlags|log.LUTC),
		warnLogger: log.New(os.Stderr, "WARN: ", log.LstdFlags|log.LUTC),
		errLogger:  log.New(os.Stderr, "ERROR: ", log.LstdFlags|log.LUTC),
		debugLogger: log.New(os.Stdout, "DEBUG: ", log.LstdFlags|log.LUTC),
		DebugMode: debug,
	}
}

func (logger *Logger) Errorf(fmtString string, args ...any) {
	logger.errLogger.Printf(fmtString, args...)
}

func (logger *Logger) Errorln(errString ...any) {
	logger.errLogger.Println(errString...)
}

func (logger *Logger) Warningf(fmtString string, args ...any) {
	logger.warnLogger.Printf(fmtString, args...)
}

func (logger *Logger) Warningln(warnString ...any) {
	logger.warnLogger.Println(warnString...)
}

func (logger *Logger) Infof(fmtString string, args ...any) {
	logger.infoLogger.Printf(fmtString, args...)
}

func (logger *Logger) Infoln(infoString ...any) {
	logger.infoLogger.Println(infoString...)
}

func (logger *Logger) Debugln(debugString ...any) {
	if logger.DebugMode {
		logger.debugLogger.Println(debugString...)
	}
}

func (logger *Logger) Debugf(fmtString string, args ...any) {
	if logger.DebugMode {
		logger.debugLogger.Printf(fmtString, args...)
	}
}