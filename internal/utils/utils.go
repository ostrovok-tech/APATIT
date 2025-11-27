package utils

import (
	"math/rand"
	"time"

	"github.com/sirupsen/logrus"
)

// RandomizedPause pauses for a random amount of time in the range [min, 2*min]
func RandomizedPause(minDuration time.Duration) {
	if minDuration == time.Duration(0) {
		return
	}
	pauseRange := minDuration.Milliseconds()
	timeToSleep := time.Duration(pauseRange+rand.Int63n(pauseRange)) * time.Millisecond
	logrus.WithField("duration", timeToSleep.String()).Debug("Pausing before next request")
	time.Sleep(timeToSleep)
}
