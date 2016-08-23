package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/emptyinterface/cloudwatchlogger"
)

func main() {

	// use app name and host as group and stream
	group := "myapp"
	stream, _ := os.Hostname()

	// Ensure AWS_REGION, AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY
	// are set in the environment for aws-sdk to utilize
	sess := session.New(nil)

	// duration is how frequently the logs are flushed to cloudwatch
	logger, err := cloudwatchlogger.NewLogger(sess, group, stream, 10*time.Second)
	if err != nil {
		log.Fatal(err)
	}

	// redirect all `log.*` calls output to cloudwatch
	log.SetOutput(logger)

	log.Println("this is logged in cloudwatch!")

	logger.WriteError(errors.New("this error sucks!"))

	start := time.Now()
	resp, err := http.Get("http://ifconfig.info")
	if err != nil {
		logger.WriteError(err)
	}

	// log the entire req/resp/duration details
	logger.WriteRoundTrip(resp, time.Since(start))

	// log arbitrary json Marshalable data
	logger.WriteJSON(map[string]interface{}{
		"stats": []int{1, 2, 3},
		"are":   "great",
	})

	// close flushes the buffer's contents before returning
	// usually deferred after NewLogger()
	logger.Close()

	// wait for log events to propagate
	time.Sleep(2 * time.Second)

	// access to underlying aws Service object allows for more specific actions
	cwresp, err := logger.Service.GetLogEvents(&cloudwatchlogs.GetLogEventsInput{
		LogGroupName:  aws.String(group),
		LogStreamName: aws.String(stream),
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	for _, event := range cwresp.Events {
		fmt.Println(event)
	}

	if _, err := logger.Service.DeleteLogGroup(&cloudwatchlogs.DeleteLogGroupInput{
		LogGroupName: aws.String(group),
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

}

// Expected output:
//
// PutLogEvents: 246.847604ms
// group or stream not found, creating...
// PutLogEvents: 104.446395ms
// {
//   IngestionTime: 1441983327848,
//   Message: "2015/09/11 10:55:26 this is logged in cloudwatch!",
//   Timestamp: 1441983326513
// }
// {
//   IngestionTime: 1441983327848,
//   Message: "{\"Type\":\"error\",\"FunctionName\":\"main.main\",\"FileName\":\"/Users/jason/go/src/github.com/jasonmoo/cloudwatchlogger/example/main.go\",\"Line\":37,\"Error\":\"this error sucks!\"}",
//   Timestamp: 1441983326513
// }
// {
//   IngestionTime: 1441983327848,
//   Message: "{\"Type\":\"roundtrip\",\"Request\":{\"Method\":\"GET\",\"URL\":{\"Scheme\":\"http\",\"Opaque\":\"\",\"User\":null,\"Host\":\"ifconfig.info\",\"Path\":\"\",\"RawPath\":\"\",\"RawQuery\":\"\",\"Fragment\":\"\"},\"Header\":{},\"ContentLength\":0},\"Response\":{\"StatusCode\":200,\"Header\":{\"Connection\":[\"keep-alive\"],\"Content-Type\":[\"text/plain\"],\"Date\":[\"Fri, 11 Sep 2015 14:55:25 GMT\"],\"Server\":[\"nginx/1.6.2\"]},\"ContentLength\":-1},\"Duration\":218847763}",
//   Timestamp: 1441983326732
// }
// {
//   IngestionTime: 1441983327848,
//   Message: "{\"are\":\"great\",\"stats\":[1,2,3]}",
//   Timestamp: 1441983326733
// }
