package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
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
var originRateLimit int

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
	originRateLimit = 0
	if os.Getenv("ORIGIN_RATE_LIMIT") != "" {
		var err error
		originRateLimit, err = strconv.Atoi(os.Getenv("ORIGIN_RATE_LIMIT"))
		if err != nil {
			panic("Invalid ORIGIN_RATE_LIMIT value")
		}
	}
}

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.Header().Set("Access-Control-Allow-Origin", "*")
		writer.Header().Set("Access-Control-Allow-Methods", "GET")

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

func isLocalOrigin(origin string) bool {
	if origin == "" {
		return false
	}
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}

	hostname := u.Hostname()

	localHostnames := []string{"localhost", "127.0.0.1", "0.0.0.0"}
	for _, lh := range localHostnames {
		if hostname == lh {
			return true
		}
	}

	return strings.HasPrefix(hostname, "192.168.")
}

func checkRateLimit(IP string, origin string) bool {
	if originRateLimit > 0 && !isLocalOrigin(origin) {
		count, _ := limiter.LoadOrStore(origin, new(int))
		*count.(*int) += 1
		return *count.(*int) < originRateLimit
	}

	count, _ := limiter.LoadOrStore(IP, new(int))
	*count.(*int) += 1
	return *count.(*int) < rateLimit
}

func getIP(request *http.Request) string {
	clientIPHeader := os.Getenv("CLIENT_IP_HEADER")
	if clientIPHeader != "" {
		return request.Header.Get(clientIPHeader)
	}

	ip, _, err := net.SplitHostPort(request.RemoteAddr)
	if err == nil {
		return ip
	}

	return request.RemoteAddr
}

func get(writer http.ResponseWriter, request *http.Request) {
	URL := request.URL.Query().Get("url")
	callback := request.URL.Query().Get("callback")

	IP := getIP(request)
	origin := request.Header.Get("Origin")

	if URL == "" {
		writer.Write([]byte("URL parameter is required."))
		return
	}

	if origin == "" {
		writer.Write([]byte("Origin header is required."))
		return
	}

	if !checkRateLimit(IP, origin) {
		rateLimitValue := rateLimit
		if originRateLimit > 0 && !isLocalOrigin(origin) {
			rateLimitValue = originRateLimit
		}
		writer.Write([]byte(fmt.Sprintf("rate limited: limit %d request (s) per minute", rateLimitValue)))
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
