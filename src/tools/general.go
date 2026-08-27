package tools

import "time"

func DefaultYear() int {
	if time.Now().Month() > time.March {
		return time.Now().Year()
	} else {
		return time.Now().Year() - 1
	}
}
