package util

import "time"

// TimestampSub 计算时间差
func TimestampSub(t1, t2 int64) time.Duration {
	unix1 := time.Unix(t1, 0)
	unix2 := time.Unix(t2, 0)

	if t1 > t2 {
		return unix1.Sub(unix2)
	}

	return unix2.Sub(unix1)
}
