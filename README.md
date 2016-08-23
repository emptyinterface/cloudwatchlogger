# CloudWatch Logger

a simple library for sending log entries to AWS CloudWatch logs.

## Embedded Use

See [full example](https://github.com/jasonmoo/cloudwatchlogger/blob/master/example/main.go).

## Agent Use

A stand-alone agent is provided for easy logging scenarios.  The [AWS CloudWatch Logs Agent](http://docs.aws.amazon.com/AmazonCloudWatch/latest/DeveloperGuide/QuickStartEC2Instance.html)
is a much more robust tool and should be considered.

	./agent usage:
	  -group string
	    	cloudwatch group name (usually app name) **required**
	  -stream string
	    	cloudwatch stream name (usually host name) (default `hostname`)
	  -flush_interval duration
	    	cloudwatch batch flushing frequency (default 30s)
	  -sock string
	    	unix socket to listen on (default "/var/run/cwagent.sock")
	  -stdin
	    	read from STDIN instead of unix socket

The agent listens on `STDIN`, or a unix socket, for any input.  Each newline terminated
line is sent to CloudWatch logs.  Any log group/stream that does not exist will be created
on first use.

	export AWS_REGION="xxx"
	export AWS_ACCESS_KEY_ID="xxx"
	export AWS_SECRET_ACCESS_KEY="xxx"

	echo "log this!" | ./agent -group junklogs -stdin

	# or as a daemon listening on unix socket
	# (using `socat` for command line socket redirection, not required)
	./agent -group junklogs -sock junk.sock &
	[1234] agent
	echo "log this too!" | socat - junk.sock

Production uses will redirect `STDOUT`/`STDERR` from running applications via
`monit`, `systemd`, `upstart`, etc.

##LICENSE
[MIT](https://github.com/jasonmoo/cloudwatchlogger/blob/master/LICENSE)