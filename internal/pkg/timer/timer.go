// NOTE: dumping some useful utils here. Might move to elsewhere or re-structure if needed
package timer

import (
	"log"
	"time"
)

func Track(msg string) (string, time.Time) {
	// track the time when this func is executed. Use with Duration.
	return msg, time.Now()
}

func Duration(msg string, start time.Time) {
	// track the duration since start. Use with Track.
	//  Example:
	//    defer Duration(Track("MyFunc"))
	log.Printf("%v: %v\n", msg, time.Since(start))
}
