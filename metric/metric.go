package metric

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

//Metric contains the channel running to store the metrics for the load testing.
type Metric struct {
	op              chan (func(*metrics))
	report_interval time.Duration
}

//Close will terminate the running metric listener.
func (m *Metric) Close() {
	close(m.op)
	return
}

//New creates a new metric listener.
func New(report_interval time.Duration) (m Metric) {
	m.op = make(chan func(*metrics))
	m.report_interval = report_interval
	go m.loop()
	return
}

//Process adds a response to the response structure.
func (m *Metric) Process(r *http.Response, et time.Duration) {
	m.op <- func(curr *metrics) {
		curr.Stat_codes[r.StatusCode]++
		curr.Total_requests++
		curr.Timer_log = append(curr.Timer_log, et)
		if r.Body != nil {
			buf, e := ioutil.ReadAll(r.Body)
			if e != nil {
				log.Printf("Metric.Process got invalid / malformed body, skipping")
				return
			}
			curr.Bytes_processed += int64(len(buf))
			r.Body.Close()
		}
		curr.Bytes_per_request = curr.Bytes_processed / curr.Total_requests
		return
	}
	return
}

//String implements the fmt.Stringer interface for the Metric structure.
func (m Metric) String() (s string) {
	sch := make(chan string)
	m.op <- func(curr *metrics) {
		msg := fmt.Sprintf("--Request Status Summary--\n")
		for code, count := range curr.Stat_codes {
			msg += fmt.Sprintf("%d: %d\n", code, count)
		}
		msg += fmt.Sprintf("Total Requests: %d\n", curr.Total_requests)
		msg += fmt.Sprintf("Total Bytes Recieved: %d\n", curr.Bytes_processed)
		msg += fmt.Sprintf("Avg Bytes Per Request: %d\n", curr.Bytes_per_request)
		msg += fmt.Sprintf("--Current State--\n")
		fmt.Println(curr)
		state, e := json.MarshalIndent(*curr, "", "    ")
		if e != nil {
			log.Printf("Error unmarshaling data.")
		}
		msg += fmt.Sprintf("%s", state)
		sch <- msg
	}
	s = <-sch
	return
}

//Marshaler creates a current json-encoding of a Metric
func (m Metric) MarshalJSON() (b []byte, e error) {
	bch := make(chan []byte)
	ech := make(chan error)
	m.op <- func(curr *metrics) {
		 tb , te := json.MarshalIndent(*curr, "", "    ")
		bch <- tb
		ech <- te
	}
	b = <- bch
	e = <- ech
	return
}
		

//loop is the primary loop for the running metric.
func (m *Metric) loop() {
	sm := &metrics{}
	sm.Stat_codes = make(map[int]int64)
	go func() {
		c := time.Tick(m.report_interval)
		for _ = range c {
			log.Printf("%s", m)
		}
	}()
	for op := range m.op {
		op(sm)
	}
}

type metrics struct {
	Stat_codes        map[int]int64
	Timer_log		  []time.Duration
	Total_requests    int64
	Bytes_processed   int64
	Bytes_per_request int64
}
