package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/emptyinterface/cloudwatchlogger"
)

var (
	hostname, _   = os.Hostname()
	stdin         = flag.Bool("stdin", false, "read from STDIN instead of unix socket")
	sock          = flag.String("sock", "/var/run/cwagent.sock", "unix socket to listen on")
	group         = flag.String("group", "", "cloudwatch group name (usually app name) **required**")
	stream        = flag.String("stream", hostname, "cloudwatch stream name (usually host name)")
	flushInterval = flag.Duration("flush_interval", 30*time.Second, "cloudwatch batch flushing frequency")
)

func init() {
	flag.Parse()

	if len(*group) == 0 {
		fmt.Printf("%s usage:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(0)
	}
}

func main() {

	log.Println("starting up")

	logger, err := cloudwatchlogger.NewLogger(session.New(nil), *group, *stream, *flushInterval)
	if err != nil {
		log.Fatal(err)
	}

	// if stdin just do the copy and close up here
	if *stdin {
		if _, err := io.Copy(logger, os.Stdin); err != nil {
			log.Println("copy err", err)
		}
		if err := logger.Close(); err != nil {
			log.Println(err)
		}
		log.Println("exiting...")
		os.Exit(0)
	}

	// in case a sock exists, remove it
	if _, err := os.Stat(*sock); err == nil {
		log.Println("found existing socket.  removing...")
		if err := os.Remove(*sock); err != nil {
			log.Fatalf("unable to remove existing socket: %q", *sock)
		}
	}

	// ensure path to socket file exists
	if err := os.MkdirAll(filepath.Dir(*sock), 0755); err != nil {
		log.Fatal(err)
	}

	// unix == SOCK_STREAM == tcp
	l, err := net.Listen("unix", *sock)
	if err != nil {
		log.Fatal(err)
	}

	// ensure we only close once
	once := sync.Once{}
	done := func() {
		if err := l.Close(); err != nil {
			log.Println(err)
		}
		if err := logger.Close(); err != nil {
			log.Println(err)
		}
		log.Println("exiting...")
		os.Exit(0)
	}
	defer once.Do(done)

	// ensure our done function triggers if the process is interrupted
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGKILL)
	go func() {
		<-sig
		once.Do(done)
	}()

	log.Println("listening on socket: ", *sock)
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("conn err", err)
			break
		}
		go func(conn net.Conn) {
			defer func() {
				if e := recover(); e != nil {
					log.Println("panic", e)
				}
			}()
			defer conn.Close()
			if _, err := io.Copy(logger, conn); err != nil {
				log.Println("copy err", err)
			}
		}(conn)
	}
}
