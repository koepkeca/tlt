package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/gorilla/context"
	"github.com/koepkeca/tlt/config"
	"github.com/koepkeca/tlt/metric"
)

const default_user_agent = "TLT - Trivial Load Tester"

func setUlimit() {
	/*
		var rLimit syscall.Rlimit
		rLimit.Max = 32768
		rLimit.Cur = 32768
		e := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
		if e != nil {
			log.Printf("Syscall: setlimit error: %s", e)
		}
	*/
	return
}

func json_writer(w http.ResponseWriter, p interface{}) {
	resp, e := json.Marshal(p)
	if e != nil {
		http.Error(w, "Internal JSON Error: ", 500)
		log.Print("JSON Error: ", e)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
	return
}

func main() {
	switch runtime.GOOS {
	case "linux":
		setUlimit()
	}
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
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token")
		if r.Method == "OPTIONS" {
			log.Println(r)
			log.Printf("XXXXXXXXXXXXXXXXXXXXXXXX")
			return
		}

		json_writer(w, m)
	})
	go func() {
		err := http.ListenAndServe(":8192", context.ClearHandler(http.DefaultServeMux))
		if err != nil {
			log.Printf("Http service disabled: %s", err)
		}
		return
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
				req_st := time.Now()
				log.Printf("Sending request %d\n", id)
				tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
				c := &http.Client{Timeout: conf.Interval, Transport: tr}
				r, e := http.NewRequest("GET", conf.Target, nil)
				if e != nil {
					m.Process(&http.Response{Status: e.Error(), StatusCode: http.StatusServiceUnavailable}, time.Since(req_st))
					log.Printf("Error creating http request: %s", e)
					return
				}
				r.Header.Set("User-Agent", default_user_agent)
				resp, e := c.Do(r)
				if e != nil {
					m.Process(&http.Response{Status: e.Error(), StatusCode: http.StatusServiceUnavailable}, time.Since(req_st))
					log.Printf("Error completing http request:%s", e)
					fmt.Println(e)
					return
				}
				//Because the process that manages the response can read it,
				//m.Process is responsible for closing the request body.
				m.Process(resp, time.Since(req_st))
				c.CloseIdleConnections()
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
