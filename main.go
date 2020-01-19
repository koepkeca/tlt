package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/koepkeca/tlt/config"
	"github.com/koepkeca/tlt/metric"
)

const default_user_agent = "TLT - Trivial Load Tester"

func main() {
	conf := config.New()
	m := metric.New(conf.Report_interval)
	go func() {
		sc := make(chan os.Signal)
		signal.Notify(sc, syscall.SIGINT)
		<-sc
		fmt.Printf("\n")
		log.Printf("[Captured Interrupt] - Summarizing and Terminating.")
		log.Printf("%s", m)
		log.Printf("Performing cleanup and terminating.")
		m.Close()
		os.Exit(0)
	}()
	tt := time.Tick(conf.Interval)
	st := time.Now()
	rand.Seed(time.Now().UTC().UnixNano())
	totalReq := int64(0)
	for _ = range tt {
		i := int64(0)
		for i < conf.Requests {
			log.Printf("Spawning request %d\n", totalReq)
			go func(id int64) {
				delay := time.Duration(rand.Int63n(conf.Interval.Milliseconds())) * time.Millisecond
				time.Sleep(delay)
				log.Printf("Sending request %d\n", id)
				tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
				c := &http.Client{Timeout: conf.Interval, Transport: tr}
				r, e := http.NewRequest("GET", conf.Target, nil)
				if e != nil {
					m.Process(&http.Response{Status: e.Error(), StatusCode: http.StatusServiceUnavailable})
					log.Printf("Error creating http request: %s", e)
					return
				}
				r.Header.Set("User-Agent", default_user_agent)
				resp, e := c.Do(r)
				if e != nil {
					m.Process(&http.Response{Status: e.Error(), StatusCode: http.StatusServiceUnavailable})
					log.Printf("Error completing http request:%s", e)
					fmt.Println(e)
					return
				}
				//Because the process that manages the response can read it,
				//m.Process is responsible for closing the request body.
				m.Process(resp)
			}(totalReq)
			i++
			totalReq++
		}
		et := time.Since(st)
		if et > conf.Runtime {
			break
		}
	}
	log.Printf("Test concluded, terminating normally. Waiting for final requests to respond...")
	time.Sleep(conf.Interval)
	log.Printf("Done.\n")
	log.Printf("%s", m)
	m.Close()
	return
}
