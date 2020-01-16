package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/koepkeca/tlt/config"
	"github.com/koepkeca/tlt/metric"
)

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
	totals := 1
	for _ = range tt {
		i := int64(0)
		for i < conf.Requests {
			log.Printf("Request %d\n",totals)
			go func() {
				c := &http.Client{}
				r, e := http.NewRequest("GET", conf.Target, nil)
				if e != nil {
					log.Printf("%s", e)
					return
				}
				resp, e := c.Do(r)
				if e != nil {
					log.Printf("%s", e)
					return
				}
				m.Process(resp)
				resp.Body.Close()
			}()
			i++
			totals++
		}
		et := time.Since(st)
		if et > conf.Runtime {
			break
		}
	}
    log.Printf("Test concluded, terminating normally.")
	time.Sleep(conf.Interval)
	log.Printf("%s",m)
	m.Close()
	return
}
