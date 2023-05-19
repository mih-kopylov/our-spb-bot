package util

import "time"

var (
	SpbLocation = time.FixedZone("UTC+3", 3*60*60)

	DefaultSendTime = time.Date(2023, time.January, 1, 5, 0, 0, 0, SpbLocation)
)
