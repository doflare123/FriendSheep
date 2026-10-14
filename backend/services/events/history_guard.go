package events

import "time"

func eventHasEnded(endTime, now time.Time) bool {
	return !endTime.IsZero() && !now.Before(endTime)
}
