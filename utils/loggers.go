package utils

import "log"

func Info(msg string) {
    log.Printf("[INFO] %s", msg)
}

func Error(msg string) {
    log.Printf("[ERROR] %s", msg)
}