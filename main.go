package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"math"
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

const timeUnavailableURL = 0

func main() {
	file, err := os.Open("sites.txt")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	res := make(map[string]float64)

	ntp := NewTransport()
	client := &http.Client{Transport: ntp}

	if readFileSuccess(file, client, res, ntp) {
		return
	}

	ticker := time.NewTicker(1 * time.Minute)
	done := make(chan bool)

	go func() {
		for {
			select {
			case <-done:
				return
			case t := <-ticker.C:
				fmt.Println("Tick at: ", t)
			}
		}
	}()
	time.Sleep(4 * time.Second)
	ticker.Stop()
	done <- true

	checkSites(client, res, ntp)

	printCheckResults(res)
}

func printCheckResults(res map[string]float64) {
	for k, v := range res {
		log.Printf("site: %s available: %v\n", k, v)
	}

	fmt.Println("Minimal duration: ", getMinimalDuration(res))
	fmt.Println("Maximal duration: ", getMaximalDuration(res))
}

func readFileSuccess(file *os.File, client *http.Client, res map[string]float64, ntp *transport) bool {
	r := bufio.NewReader(file)
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Fatalf("read file line error: %v", err)
			return true
		}
		line = getURL(line)
		res[line] = timeUnavailableURL
	}

	return false
}

func checkSites(client *http.Client, sites map[string]float64, ntp *transport) {
	for url := range sites {
		resp, err := client.Get(url)
		if err != nil {
			log.Printf("get site error: %v", err)
			continue
		}
		resp.Body.Close()

		sites[url] = ntp.Duration().Seconds()
		fmt.Printf("URL: %s duration:%f sec\n", url, ntp.Duration().Seconds())
	}
}

func getURL(line string) string {
	return fmt.Sprintf("https://www.%s", line[:len(line)-1])
}

func NewTransport() *transport {
	t := &transport{
		dialer: &net.Dialer{
			Timeout:   15 * time.Second,
			KeepAlive: 15 * time.Second,
		},
	}
	t.rt = &http.Transport{
		Dial:                t.dial,
		Proxy:               http.ProxyFromEnvironment,
		TLSHandshakeTimeout: 10 * time.Second,
	}

	return t
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

func (t *transport) dial(network, addr string) (net.Conn, error) {
	t.connStart = time.Now()
	conn, err := t.dialer.Dial(network, addr)
	if err != nil {
		return nil, err
	}
	t.connEnd = time.Now()

	return conn, nil
}

func getMinimalDuration(m map[string]float64) float64 {
	min := math.MaxFloat64
	for _, v := range m {
		if v != timeUnavailableURL && v < min {
			min = v
		}
	}

	return min
}
func getMaximalDuration(m map[string]float64) float64 {
	max := float64(0)
	for _, v := range m {
		if v > max {
			max = v
		}
	}

	return max
}
