package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

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
var rateLimit int

func init() {
	client = &http.Client{}
	limiter = sync.Map{}
	rateLimit = 30
	if os.Getenv("RATE_LIMIT") != "" {
		var err error
		rateLimit, err = strconv.Atoi(os.Getenv("RATE_LIMIT"))
		if err != nil {
			panic("Invalid RATE_LIMIT value")
		}
	}
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

func tunnel(URL string) Response {
	request, err := http.NewRequest("GET", URL, nil)

	if err != nil {
		return Response{
			Status: Status{
				URL:  URL,
				Code: 500,
			},
		}
	}

	response, err := client.Do(request)

	if err != nil {
		return Response{
			Status: Status{
				URL:  URL,
				Code: 500,
			},
		}
	}

	plain, err := ioutil.ReadAll(response.Body)

	if err != nil {
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

	return result
}

func check(address string) bool {
	var value int

	count, _ := limiter.LoadOrStore(address, &value)

	*count.(*int) += 1

	return *count.(*int) < rateLimit
}

func getIP(request *http.Request) string {
	clientIPHeader := os.Getenv("CLIENT_IP_HEADER")
	if clientIPHeader != "" {
		return request.Header.Get(clientIPHeader)
	}

	return request.RemoteAddr
}

func get(writer http.ResponseWriter, request *http.Request) {
	URL := request.URL.Query().Get("url")

	if URL == "" {
		writer.Write([]byte("URL parameter is required."))
		return
	}

	callback := request.URL.Query().Get("callback")

	IP := getIP(request)
	allowed := check(IP)

	if !allowed {
		writer.Write([]byte(fmt.Sprintf("rate limited: you have a max of %d request (s) per minute", rateLimit)))
		return
	}

	body, _ := json.Marshal(tunnel(URL))

	if callback != "" {
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
