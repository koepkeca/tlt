package main

import (
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
			log.Printf("Spawning request id %d", totalReq)
			req := &Request{Id: totalReq, Conf: conf, Metric: m}
			go req.Exec()
			i++
			totalReq++
		}
		et := time.Since(st)
		if et > conf.Runtime {
			break
		}
	}
	log.Printf("Test concluded, terminating normally. Waiting for final requests to complete...")
	time.Sleep(conf.Interval)
	log.Printf("Done.\n")
	log.Printf("%s", m)
	ofp, e := os.Create(conf.Outfile)
	if e != nil {
		log.Printf("%s", e)
		return
	}
	_, e = fmt.Fprintf(ofp, "%s", m)
	if e != nil {

		log.Printf("Unable to write session result: %s", e)
	}
	m.Close()
	return
}
