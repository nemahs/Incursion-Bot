package logging

import (
	"log"
	"os"
)

type Logger struct {
	infoLogger *log.Logger
	warnLogger *log.Logger
	errLogger  *log.Logger
}

func NewLogger() Logger {
	return Logger{
		infoLogger: log.New(os.Stdout, "INFO: ", log.LstdFlags|log.LUTC),
		warnLogger: log.New(os.Stdout, "WARN: ", log.LstdFlags|log.LUTC),
		errLogger:  log.New(os.Stdout, "ERROR: ", log.LstdFlags|log.LUTC),
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