package util

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTime_TimestampSub(t *testing.T) {
	cases := []struct {
		name         string
		t1           int64
		t2           int64
		expectHour   float64
		expectMinute float64
		expectSecond float64
	}{
		{
			name:         "t1 < t2",
			t1:           1721617200, // 2024-07-22 11:00:00
			t2:           1721622600, // 2024-07-22 12:30:10
			expectHour:   1.5,
			expectMinute: 90,
			expectSecond: 5400,
		},
		{
			name:         "t1 > t2",
			t1:           1721622600, // 2024-07-22 12:30:10
			t2:           1721617200, // 2024-07-22 11:00:00
			expectHour:   1.5,
			expectMinute: 90,
			expectSecond: 5400,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			sub := TimestampSub(c.t1, c.t2)
			assert.Equal(t, c.expectHour, sub.Hours())
			assert.Equal(t, c.expectMinute, sub.Minutes())
			assert.Equal(t, c.expectSecond, sub.Seconds())
		})
	}
}
