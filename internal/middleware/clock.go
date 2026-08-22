package middleware

import "time"

// 테스트에서 시계를 교체할 수 있게 분리한다.
var (
	timeNow   = time.Now
	timeSince = func(t time.Time) float64 { return time.Since(t).Seconds() }
)
