package main

import (
	"crypto/tls"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/koepkeca/tlt/config"
	"github.com/koepkeca/tlt/metric"
)

// Request contains the configuration data and telemetry runner required for executing the
// HTTP request and capturing the response data.
type Request struct {
	Id     int64
	Conf   config.C
	Metric metric.Metric
}

func (r Request) Exec() {
	log.Printf("Executing request %d\n", r.Id)
	for _, nextTarget := range r.Conf.Target {
		req_st := time.Now()
		log.Printf("Request %d, Target %s", r.Id, nextTarget)
		tr := &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		c := &http.Client{Timeout: r.Conf.Interval, Transport: tr}
		send, e := http.NewRequest("GET", nextTarget, nil)
		if e != nil {
			r.Metric.Process(r.Id, nextTarget, &http.Response{Status: e.Error(), StatusCode: http.StatusServiceUnavailable}, req_st)
			log.Printf("Error creating http request: %s", e)
			continue
		}
		send.Header.Set("User-Agent", default_user_agent)
		resp, e := c.Do(send)
		if e != nil {
			r.Metric.Process(r.Id, nextTarget, &http.Response{Status: e.Error(), StatusCode: http.StatusServiceUnavailable}, req_st)
			log.Printf("Error completing http request: (%s) [%s]", nextTarget, e)
			log.Println(e)
			continue
		}
		//Because the process that manages the response can read it,
		//m.Process is responsible for closing the request body.
		r.Metric.Process(r.Id, nextTarget, resp, req_st)
		delay := time.Duration(rand.Int63n(r.Conf.Interval.Milliseconds())) * time.Millisecond
		ct := time.Now()
		execTime := time.Duration(ct.UnixMilli()-time.Since(req_st).Milliseconds()) * time.Millisecond
		time.Sleep(delay - execTime)
		c.CloseIdleConnections()
	}
	return
}
