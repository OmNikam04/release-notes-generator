package logger

import (
    "os"
    "github.com/rs/zerolog"
    "github.com/rs/zerolog/log"
)

var Logger zerolog.Logger

func Init(env string) {
    if env == "development" {
        // Pretty console output for development
        Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
            With().
            Timestamp().
            Caller().
            Logger()
    } else {
        // JSON output for production
        Logger = zerolog.New(os.Stdout).
            With().
            Timestamp().
            Caller().
            Logger()
    }
    
    log.Logger = Logger
}

func Info() *zerolog.Event {
    return Logger.Info()
}

func Error() *zerolog.Event {
    return Logger.Error()
}

func Debug() *zerolog.Event {
    return Logger.Debug()
}

func Warn() *zerolog.Event {
    return Logger.Warn()
}