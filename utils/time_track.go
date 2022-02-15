package utils

import (
	"log"
	"time"
)

func TimeTrack(start time.Time, logName string) {
	elapsed := time.Since(start)
	log.Printf("func: %s, cost: %s\n", logName, elapsed)
}
