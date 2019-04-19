package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/gorilla/rpc/json"
	"log"
	"net/http"
	"os"
	"syscall"
	"time"
)

var (
	nzbGetUser string = os.Getenv("NZBGET_USER")
	nzbGetPass string = os.Getenv("NZBGET_PASS")
	pauseThreshold = flag.Uint64("pause-threshold", 10, "pause threshold in GB" )
	unpauseThreshold = flag.Uint64("unpause-threshold", 50, "unpause threshold in GB" )
	ssl = flag.Bool("ssl", false, "Use SSL for communication with nzbget" )
	hostname = flag.String("host", "localhost", "hostname for nzbget" )
	port = flag.String("port", "6789", "port for nzbget" )
	dir = flag.String("dir", "/", "directory to check" )
	paused bool
)

func main() {
	flag.Parse()
	if len(nzbGetUser) == 0 {
		log.Fatalf("missing env NZBGET_USER\n")
	}
	if len(nzbGetPass) == 0 {
		log.Fatalf("missing env NZBGET_USER\n")
	}
	for {
		var resp bool
		s, err := getDiskSpace(*dir)
		if err != nil {
			log.Fatalf("Failed to determine disk space: %v\n", err)
		}
		if s < (*pauseThreshold) && !paused {
			fmt.Printf("Pausing Downloads: current free disk space is: %v GB. Threshold to pause (%v GB) reached.\n", s, *pauseThreshold)
			if resp, err = nzbGet("pausedownload"); err != nil {
				log.Fatalf("Error Pausing Downloads: %v\n", err)
			}
			paused = true
			fmt.Println(resp)
		} else if s > (*unpauseThreshold) && paused {
			fmt.Printf("Resuming downloads: current free disk space is: %v GB. Threshold to unpause (%v GB) reached.\n", s, *unpauseThreshold)
			if resp, err = nzbGet("resumedownload"); err != nil {
				log.Fatalf("Error Pausing Downloads: %v\n", err)
			}
			paused = false
			fmt.Println(resp)
		}
		if paused {
			fmt.Printf("Downloads Paused, current free disk space is: %v GB. Threshold to unpause (%v GB) has not been reached.\n", s, *unpauseThreshold)
		} else {
			fmt.Printf("Nothing to see here, current free disk space is: %v GB. Threshold to pause (%v GB) has not been reached.\n", s, *pauseThreshold)
		}
		time.Sleep(time.Second * 60)
	}

}

func getDiskSpace(path string) (uint64, error) {
	var stat syscall.Statfs_t

	err := syscall.Statfs(path, &stat)
	if err != nil {
		return 0, err
	}

	return (stat.Bavail * uint64(stat.Bsize)) / 1024 / 1024 / 1024, nil
}

func nzbGet(action string) (bool, error) {
	var scheme string

	if *ssl {
		scheme = "https"
	} else {
		scheme = "http"
	}

	url := fmt.Sprintf("%s://%s:%s/%s:%s/jsonrpc", scheme, *hostname, *port, nzbGetUser, nzbGetPass)
	message, err := json.EncodeClientRequest(action, nil)
	if err != nil {
		return false, fmt.Errorf("error encoding client request: %v", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(message))
	if err != nil {
		return false, fmt.Errorf("error creating request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	client := new(http.Client)
	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("error sending request: %v", err)
	}

	defer func () {
		err := resp.Body.Close()
		if err != nil {
			log.Fatalf("Error closing response body %v", err)
		}
	}()

	var result bool
	err = json.DecodeClientResponse(resp.Body, &result)
	if err != nil {
		return false, fmt.Errorf("error decoding client response: %v", err)
	}
	return result, nil
}
