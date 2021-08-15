package util

import "time"

func NowInt() int {
	return int(time.Now().Unix())
}
