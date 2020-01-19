package metric

import (
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
func (m *Metric) Process(r *http.Response) {
	m.op <- func(curr *metrics) {
		curr.stat_codes[r.StatusCode]++
		curr.total_requests++
		if r.Body != nil {
			buf, e := ioutil.ReadAll(r.Body)
			if e != nil {
				log.Printf("Metric.Process got invalid / malformed body, skipping")
				return
			}
			curr.bytes_processed += int64(len(buf))
			r.Body.Close()
		}
		curr.bytes_per_request = curr.bytes_processed / curr.total_requests
		return
	}
	return
}

//String implements the fmt.Stringer interface for the Metric structure.
func (m Metric) String() (s string) {
	sch := make(chan string)
	m.op <- func(curr *metrics) {
		msg := fmt.Sprintf("--Request Status Summary--\n")
		for code, count := range curr.stat_codes {
			msg += fmt.Sprintf("%d: %d\n", code, count)
		}
		msg += fmt.Sprintf("Total Requests: %d\n", curr.total_requests)
		msg += fmt.Sprintf("Total Bytes Recieved: %d\n", curr.bytes_processed)
		msg += fmt.Sprintf("Avg Bytes Per Request: %d\n", curr.bytes_per_request)
		sch <- msg
	}
	s = <-sch
	return
}

//loop is the primary loop for the running metric.
func (m *Metric) loop() {
	sm := &metrics{}
	sm.stat_codes = make(map[int]int64)
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
	stat_codes        map[int]int64
	total_requests    int64
	bytes_processed   int64
	bytes_per_request int64
}
