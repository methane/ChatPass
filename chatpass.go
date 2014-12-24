package main

import (
	"flag"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var regPostMessage *regexp.Regexp = regexp.MustCompile("/rooms/([^/]+)/messages")

type postRequest struct {
	room, body string
}

var reqCh chan postRequest = make(chan postRequest, 10000)

func roomHandler(w http.ResponseWriter, r *http.Request) {
	x := regPostMessage.FindSubmatch([]byte(r.URL.Path))
	if len(x) == 0 {
		log.Println("Bad URL: ", r.URL.String())
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	room := string(x[1])

	body := r.FormValue("body")
	if body == "" {
		log.Println("Empty body.  URL=%s, remote=%s", r.URL, r.RemoteAddr)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Printf("request: room=%q, body=%q", room, body)

	reqCh <- postRequest{room, body}
}

func sender(token string) {
	for req := range reqCh {
		log.Printf("Sending %#v", req)
		val := url.Values(make(map[string][]string))
		val.Set("body", req.body)
		body := val.Encode()
		url := "https://api.chatwork.com/v1/rooms/" + req.room + "/messages"
		for {
			req, err := http.NewRequest("POST", url, strings.NewReader(body))
			if err != nil {
				log.Println(err)
				continue
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			req.Header.Set("X-ChatWorkToken", token)
			res, err := http.DefaultClient.Do(req)

			if err != nil {
				log.Println(err)
			}
			if res == nil {
				break
			}
			if res.StatusCode == 429 {
				resetStr := res.Header.Get("X-RateLimit-Reset")
				resetAt := time.Now().Add(time.Second * 30)
				if resetStr != "" {
					resetInt, err := strconv.Atoi(resetStr)
					if err != nil {
						resetAt = time.Unix(int64(resetInt), 0)
					}
				}
				time.Sleep(resetAt.Sub(time.Now()) + time.Second*5)
				continue
			}
			break
		}
	}
}

func main() {
	var token string
	var addr string
	flag.StringVar(&token, "token", "", "ChatWork API Token")
	flag.StringVar(&addr, "addr", ":8000", "Address to listen")
	flag.Parse()

	if token == "" {
		log.Fatal("token is required.")
	}

	go sender(token)
	http.HandleFunc("/rooms/", roomHandler)
	log.Fatal(http.ListenAndServe(addr, nil))
}
