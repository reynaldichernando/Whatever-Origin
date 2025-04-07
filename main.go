package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/honeycombio/libhoney-go"
)

const RATE_LIMIT = 600

type Status struct {
	URL  string `json:"url"`
	Type string `json:"content_type"`
	Code int    `json:"http_code"`
}

type Response struct {
	Content string `json:"contents"`
	Status  Status `json:"status"`
}

var client *http.Client
var limiter sync.Map

var getLogger *libhoney.Client

var bypass = "Fly-Client-Ip"

func init() {
	var err error

	client = &http.Client{}

	key := os.Getenv("HONEY")

	bypassKey := os.Getenv("BYPASS")

	if bypassKey != "" {
		bypass = bypassKey
	}

	getLogger, err = libhoney.NewClient(libhoney.ClientConfig{
		APIKey:  key,
		Dataset: "whatever-tunnel",
	})

	if err != nil {
		panic(err)
	}

	commit, err := ioutil.ReadFile(".git/refs/heads/master")

	if err != nil {
		panic(err)
	}

	getLogger.AddField("git", strings.TrimSpace(string(commit)))
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Allow", "*")
		writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, PUT, DELETE, OPTIONS")
		writer.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Referer, User-Agent")

		next.ServeHTTP(writer, request)
	})
}

func tunnel(URL string, event *libhoney.Event) Response {
	event.AddField("origin", URL)

	request, err := http.NewRequest("GET", URL, nil)

	if err != nil {
		event.AddField("error", "invalid request")
		event.AddField("error_raw", err)

		return Response{
			Status: Status{
				URL:  URL,
				Code: 500,
			},
		}
	}

	response, err := client.Do(request)

	if err != nil {
		event.AddField("error", "request failed")
		event.AddField("error_raw", err)

		return Response{
			Status: Status{
				URL:  URL,
				Code: 500,
			},
		}
	}

	plain, err := ioutil.ReadAll(response.Body)

	if err != nil {
		event.AddField("error", "failed to read body")
		event.AddField("error_raw", err)

		return Response{
			Status: Status{
				URL:  URL,
				Code: 500,
			},
		}
	}

	result := Response{
		Content: string(plain),
		Status: Status{
			URL:  URL,
			Type: response.Header.Get("Content-Type"),
			Code: response.StatusCode,
		},
	}

	event.AddField("content", result.Content)
	event.AddField("status_url", result.Status.URL)
	event.AddField("status_type", result.Status.Type)
	event.AddField("status_code", result.Status.Code)

	return result
}

func check(address string) bool {
	var value int

	count, _ := limiter.LoadOrStore(address, &value)

	*count.(*int) += 1

	return *count.(*int) < RATE_LIMIT
}

func get(writer http.ResponseWriter, request *http.Request) {
	event := getLogger.NewEvent()
	defer event.Send()

	start := time.Now()

	defer func(event *libhoney.Event) {
		event.AddField("duration", time.Since(start).Milliseconds())
	}(event)

	URL := request.URL.Query().Get("url")
	IP := request.Header.Get(bypass)

	event.AddField("url", URL)
	event.AddField("ip", IP)

	if URL == "" {
		event.AddField("error", "failed to find parameter")

		writer.Write([]byte("URL parameter is required."))
		return
	}

	callback := request.URL.Query().Get("callback")

	allowed := check(IP)

	if allowed == false {
		event.AddField("error", "rate limited")

		writer.Write([]byte(fmt.Sprintf("rate limited: you have a max of %d request per second", RATE_LIMIT)))
		return
	}

	body, _ := json.Marshal(tunnel(URL, event))

	if callback != "" {
		event.AddField("callback", callback)

		writer.Header().Set("Content-Type", "application/x-javascript")
		body = []byte(callback + "(" + string(body) + ")")
	} else {
		writer.Header().Set("Content-Type", "application/json")
	}

	writer.Write(body)
}

func main() {
	go func() {
		for {
			time.Sleep(1 * time.Minute)

			limiter = sync.Map{}
		}
	}()

	http.Handle("/get", CORS(http.HandlerFunc(get)))
	http.Handle("/", http.FileServer(http.Dir("./static")))

	panic(http.ListenAndServe(":8080", nil))
}
