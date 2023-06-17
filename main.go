package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

type transport struct {
	dialer    *net.Dialer
	rt        http.RoundTripper
	connStart time.Time
	connEnd   time.Time
	reqStart  time.Time
	reqEnd    time.Time
}

func main() {
	file, err := os.Open("sites.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	res := make(map[string]float64)

	ntp := NewTransport()
	client := &http.Client{Transport: ntp}

	r := bufio.NewReader(file)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Fatalf("read file line error: %v", err)
			return
		}
		line = cutEndOfLine(line)

		resp, err := client.Get(line)
		if err != nil {
			log.Fatalf("get url error: %v", err)
		}
		resp.Body.Close()

		res[line] = ntp.Duration().Seconds()
		fmt.Printf("URL: %s duration:%f sec\n", line, ntp.Duration().Seconds())
		fmt.Printf("URL: %s connection duration:%f sec\n", line, ntp.ConnDuration().Seconds())
		fmt.Printf("URL: %s request duration:%f sec\n", line, ntp.ReqDuration().Seconds())

	}

	for k, v := range res {
		log.Printf("site: %s available: %v\n", k, v)
	}
}

func cutEndOfLine(line string) string {
	return line[:len(line)-1]
}

func ping(url string) bool {
	timeout := 3 * time.Second
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:80", url), timeout)
	if err != nil {
		log.Println("site unreachable error: ", err)
		return false
	}

	defer conn.Close()

	return true
}

func NewTransport() *transport {
	return &transport{
		dialer: &net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 15 * time.Second,
		},
		rt: &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}
}

func (t *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.reqStart = time.Now()
	response, err := t.rt.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	t.reqEnd = time.Now()

	return response, nil
}

func (t *transport) Duration() time.Duration {
	return t.reqEnd.Sub(t.reqStart)
}

func (t *transport) ConnDuration() time.Duration {
	return t.connEnd.Sub(t.connStart)
}

func (t *transport) ReqDuration() time.Duration {
	return t.Duration() - t.ConnDuration()
}
