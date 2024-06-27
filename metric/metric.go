package metric

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

var Err_Timeout = errors.New("context deadline exceeded (Client.Timeout or context cancellation while reading body)")

// Metric contains the channel running to store the metrics for the load testing.
type Metric struct {
	op              chan (func(*metrics))
	report_interval time.Duration
}

// Close will terminate the running metric listener.
func (m *Metric) Close() {
	close(m.op)
	return
}

// New creates a new metric listener.
func New(report_interval time.Duration) (m Metric) {
	m.op = make(chan func(*metrics))
	m.report_interval = report_interval
	go m.loop()
	return
}

// Process adds a response to the response structure.
func (m *Metric) Process(id int64, target string, r *http.Response, st time.Time) {
	m.op <- func(curr *metrics) {
		ttfb := time.Since(st)
		if r.Body != nil {
			buf, e := io.ReadAll(r.Body)
			if e != nil {
				if e.Error() == Err_Timeout.Error() {
					r.StatusCode = http.StatusRequestTimeout
					log.Printf("Metric.Process got timeout %s", time.Since(st))
				} else {
					log.Printf("Metric.Process got invalid / malformed body, skipping, [%s]", e)
					log.Println("r.Body: ", r.Body)
					r.StatusCode = http.StatusServiceUnavailable
				}
			}
			stats := RespMeta{Id: id,
				Ttfb:            ttfb,
				Total:           time.Since(st),
				Url:             target,
				Response:        r.StatusCode,
				Bytes_processed: int64(len(buf)),
			}
			curr.Report_log[target] = append(curr.Report_log[target], stats)
			curr.Bytes_processed += int64(len(buf))
			r.Body.Close()
		}
		curr.Stat_codes[r.StatusCode]++
		curr.Total_requests++
		curr.Bytes_per_request = curr.Bytes_processed / curr.Total_requests
		return
	}
	return
}

// String implements the fmt.Stringer interface for the Metric structure.
func (m Metric) String() (s string) {
	sch := make(chan string)
	m.op <- func(curr *metrics) {
		state, e := json.MarshalIndent(*curr, "", "    ")
		if e != nil {
			log.Printf("Error unmarshaling data.")
		}
		msg := fmt.Sprintf("%s", state)
		sch <- msg
	}
	s = <-sch
	return
}

// Marshaler creates a current json-encoding of a Metric
func (m Metric) MarshalJSON() (b []byte, e error) {
	bch := make(chan []byte)
	ech := make(chan error)
	m.op <- func(curr *metrics) {
		tb, te := json.MarshalIndent(*curr, "", "    ")
		bch <- tb
		ech <- te
	}
	b = <-bch
	e = <-ech
	return
}

// loop is the primary loop for the running metric.
func (m *Metric) loop() {
	sm := &metrics{}
	sm.Stat_codes = make(map[int]int64)
	sm.Report_log = make(map[string][]RespMeta)
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
	Report_log        map[string][]RespMeta
	Total_requests    int64
	Bytes_processed   int64
	Bytes_per_request int64
}

type RespMeta struct {
	Id              int64
	Ttfb            time.Duration
	Total           time.Duration
	Url             string
	Response        int
	Bytes_processed int64
}
