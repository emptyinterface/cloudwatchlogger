package cloudwatchlogger

import (
	"testing"
	"time"
)

func TestReady(t *testing.T) {

	var ct int

	r := NewRateLimiter(3, time.Second)

	// expecting 3 in first second + 3 in next 20ms
	const expected = 6
	end := time.Now().Add(time.Second + 20*time.Millisecond)
	for r.Ready() && time.Now().Before(end) {
		ct++
	}

	if ct != expected {
		t.Errorf("Expected %d executions, got %d", expected, ct)
	}

}
