package main

import (
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

const default_user_agent = "TLT - Trivial Load Tester github.com/koepkeca/tlt"

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
				c := &http.Client{}
				r, e := http.NewRequest("GET", conf.Target, nil)
				if e != nil {
					log.Printf("Error creating http request: %s", e)
					return
				}
				r.Header.Set("User-Agent", default_user_agent)
				resp, e := c.Do(r)
				if e != nil {
					log.Printf("Error performing http request:%s", e)
					return
				}
				m.Process(resp)
				resp.Body.Close()
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
