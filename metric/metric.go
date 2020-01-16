package metric

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

//Metric contains the channel running to store the metrics for the load testing.
type Metric struct {
	op        chan (func(*metrics))
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
		return
	}
	return
}

//String implements the fmt.Stringer interface for the Metric structure.
func (m Metric) String() (s string) {
	sch := make(chan string)
	m.op <- func(curr *metrics) {
		tmp := fmt.Sprintf("--Request Status Summary--\n")
		for code, count := range curr.stat_codes {
			tmp += fmt.Sprintf("%d: %d\n", code, count)
		}
		sch <- tmp
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
	stat_codes map[int]int64
}
