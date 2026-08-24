package middleware

import "time"

// 테스트에서 시계를 교체할 수 있게 분리한다.
// 패키지 전역 변수이므로 병렬 테스트에서 동시 스왑하지 말 것.
var (
	timeNow   = func() time.Time { return time.Now() }
	timeSince = func(t time.Time) time.Duration { return time.Since(t) }
)
