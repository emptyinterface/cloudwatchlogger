package cloudwatchlogger

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

var (
	group  = "test_group"
	stream = "test_stream"
	sess   = session.New(nil)
)

func TestLogger(t *testing.T) {

	logger, err := NewLogger(sess, group, stream, time.Second)
	if err != nil {
		t.Error(err)
	}

	defer func() {
		// manually delete group when done
		_, err = logger.Service.DeleteLogGroup(&cloudwatchlogs.DeleteLogGroupInput{
			LogGroupName: &group,
		})
		if err != nil {
			t.Error(err)
		}
	}()

	var testData = map[string]int64{
		"ts":         time.Now().UTC().UnixNano(),
		"so":         1,
		"many":       2,
		"interfaces": 3,
	}

	if err := logger.WriteJSON(testData); err != nil {
		t.Error(err)
	}

	if err := logger.Close(); err != nil {
		t.Error(err)
	}

	// wait for the log entry to become avail
	fmt.Println("waiting 5 sec for log entry to propagate")
	time.Sleep(5 * time.Second)

	resp, err := logger.Service.GetLogEvents(&cloudwatchlogs.GetLogEventsInput{
		// // A point in time expressed as the number of milliseconds since Jan 1, 1970
		// // 00:00:00 UTC.
		// StartTime *int64 `locationName:"startTime" type:"long"`
		// // A point in time expressed as the number of milliseconds since Jan 1, 1970
		// // 00:00:00 UTC.
		// EndTime *int64 `locationName:"endTime" type:"long"`

		LogGroupName:  &group,
		LogStreamName: &stream,
	})

	if err != nil {
		t.Error(err)
	}

	if len(resp.Events) != 1 {
		t.Errorf("Expected 1 event, got %d", len(resp.Events))
	}
	event := resp.Events[0]
	if event.Timestamp == nil {
		t.Error("Expected non-nil Timestamp")
	}
	if event.Message == nil {
		t.Error("Expected non-nil Message")
	}
	var respData map[string]int64
	if err := json.NewDecoder(strings.NewReader(*event.Message)).Decode(&respData); err != nil {
		t.Error(err)
	}
	for key, val := range testData {
		if v := respData[key]; v != val {
			t.Errorf("Expected %v: %v, got %v", key, val, v)
		}
	}

}

func BenchmarkLoggerWrite(b *testing.B) {

	// set flush time at 1m and don't close, just exit test func
	// should allow for write only benchmarks, without filling test log
	logger, err := NewLogger(sess, group, stream, time.Minute)
	if err != nil {
		b.Error(err)
	}

	msg := []byte(fmt.Sprintf("%s: 'this is my log entry'\n", time.Now()))

	b.SetBytes(int64(len(msg)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Write(msg)
	}

}

func BenchmarkLoggerWriteJSON(b *testing.B) {

	// set flush time at 1m and don't close, just exit test func
	// should allow for write only benchmarks, without filling test log
	logger, err := NewLogger(sess, group, stream, time.Minute)
	if err != nil {
		b.Error(err)
	}

	type TestEntry struct {
		Ts    time.Time
		Entry string
	}

	msg := &TestEntry{time.Now(), "this is my entry"}

	data, _ := json.Marshal(msg)

	b.SetBytes(int64(len(data)))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.WriteJSON(msg)
	}

}
