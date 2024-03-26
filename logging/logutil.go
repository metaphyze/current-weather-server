package logging

import (
	"fmt"
	"github.com/sirupsen/logrus"
	easy "github.com/t-tomalak/logrus-easy-formatter"
	"io"
	"os"
	"sync"
	"time"
)

// Whether or not logging should only be to file or also to standard out.
var LOG_TO_STDOUT_ALSO = true

// The prefix to use for log files
var log_file_prefix = ""

// The directory where log files will be stored
var log_file_dir = ""

// Mutex used when creating a new log file
var logFileMutex sync.Mutex

// When the log file name changes (because its based on the current date)
// we use this variable to detect the change.
var previousLogFileName = ""

// The file where logs are currently being stored
var logFile os.File

const LoggerDateFormat = "2006-01-02 15:04:05.000"

// The log file prefix is limited to characters a-z, A-Z, and 0-9
func validateLogFilePrefix(prefix string) error {
	if prefix == "" {
		return fmt.Errorf("Empty log file prefix")
	}

	for _, ch := range prefix {
		if !(ch >= 'a' && ch <= 'z' ||
			ch >= 'A' && ch <= 'Z' ||
			ch >= '0' && ch <= '9') {
			return fmt.Errorf("File prefixes must only contain a-z, A-Z, or 0-9 characters")
		}
	}

	return nil
}

// Initialize the logger with the file prefix and the log directory
func Initialize(logFilePrefix, logFileDir string) error {
	err := validateLogFilePrefix(logFilePrefix)

	if err != nil {
		return err
	}

	log_file_prefix = logFilePrefix

	dir, err := os.Stat(logFileDir)

	if err != nil {
		return err
	}

	if !dir.IsDir() {
		return fmt.Errorf("%v is not a directory", logFileDir)
	}

	log_file_dir = logFileDir
	formatter := &easy.Formatter{
		TimestampFormat: LoggerDateFormat,
		LogFormat:       "[%lvl%] %time% %msg%\n",
	}

	logrus.SetFormatter(formatter)
	return nil
}

// The file name is based on the date so when the date changes, the file name changes
func getCurrentLogFileName() string {
	return log_file_prefix + "." + time.Now().Format("2006.01.02") + ".log"
}

func LogInfo(requestId uint64, message string) {
	logToFile(requestId, message, logrus.InfoLevel)
}

func LogError(requestId uint64, message string) {
	logToFile(requestId, message, logrus.ErrorLevel)
}

func LogHTTPError(requestId uint64, message string, statusCode int) {
	logToFile(requestId, fmt.Sprintf("[HTTP status: %v] %v", statusCode, message), logrus.ErrorLevel)
}

func LogDebug(requestId uint64, message string) {
	logToFile(requestId, message, logrus.DebugLevel)
}

func LogWarn(requestId uint64, message string) {
	logToFile(requestId, message, logrus.WarnLevel)
}

func LogPanic(requestId uint64, message string) {
	logToFile(requestId, message, logrus.PanicLevel)
}

func LogFatal(requestId uint64, message string) {
	logToFile(requestId, message, logrus.FatalLevel)
}

func LogTrace(requestId uint64, message string) {
	logToFile(requestId, message, logrus.TraceLevel)
}

func logToFile(requestId uint64, message string, level logrus.Level) {
	if log_file_dir == "" {
		return
	}

	logFileMutex.Lock()
	defer logFileMutex.Unlock()
	currentLogFile := getCurrentLogFileName()

	// If the log file name has changed, we need to close the current logFile and
	// open a new one with the new file name.
	if currentLogFile != previousLogFileName {
		logFile.Close()
		fullPath := log_file_dir + string(os.PathSeparator) + currentLogFile
		f, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err == nil {
			logFile = *f
			previousLogFileName = currentLogFile

			if LOG_TO_STDOUT_ALSO {
				mw := io.MultiWriter(os.Stdout, f)
				logrus.SetOutput(mw)
			} else {
				logrus.SetOutput(f)
			}
		} else {
			fmt.Printf("Error opening log file (%v): %v", fullPath, err)
			return
		}
	}

	logLine := fmt.Sprintf("[requestNum=%v] %v", requestId, message)

	switch level {
	case logrus.InfoLevel:
		logrus.Infoln(logLine)
	case logrus.ErrorLevel:
		logrus.Errorln(logLine)
	case logrus.DebugLevel:
		logrus.Debugln(logLine)
	case logrus.FatalLevel:
		logrus.Fatalln(logLine)
	case logrus.TraceLevel:
		logrus.Traceln(logLine)
	case logrus.PanicLevel:
		logrus.Panicln(logLine)
	case logrus.WarnLevel:
		logrus.Warningln(logLine)
	default:
		logrus.Infoln(logLine)
	}

}
