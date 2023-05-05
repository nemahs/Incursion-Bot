package logging

import (
	"fmt"
	"log"
	"os"
)

type Logger struct {
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errLogger   *log.Logger
	debugLogger *log.Logger
	DebugMode   bool
}

var logs Logger

func InitLogger(debug bool) {

	if debug {
		fmt.Printf("Logger initialized in debug mode.")
	}

	logs.infoLogger = log.New(os.Stdout, "INFO: ", log.LstdFlags|log.LUTC)
	logs.warnLogger = log.New(os.Stderr, "WARN: ", log.LstdFlags|log.LUTC)
	logs.errLogger = log.New(os.Stderr, "ERROR: ", log.LstdFlags|log.LUTC)
	logs.debugLogger = log.New(os.Stdout, "DEBUG: ", log.LstdFlags|log.LUTC)
	logs.DebugMode = debug
}

func Errorf(fmtString string, args ...any) {
	logs.errLogger.Printf(fmtString, args...)
}

func Errorln(errString ...any) {
	logs.errLogger.Println(errString...)
}

func Warningf(fmtString string, args ...any) {
	logs.warnLogger.Printf(fmtString, args...)
}

func Warningln(warnString ...any) {
	logs.warnLogger.Println(warnString...)
}

func Infof(fmtString string, args ...any) {
	logs.infoLogger.Printf(fmtString, args...)
}

func Infoln(infoString ...any) {
	logs.infoLogger.Println(infoString...)
}

func Debugln(debugString ...any) {
	if logs.DebugMode {
		logs.debugLogger.Println(debugString...)
	}
}

func Debugf(fmtString string, args ...any) {
	if logs.DebugMode {
		logs.debugLogger.Printf(fmtString, args...)
	}
}
