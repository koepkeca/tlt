package config

import (
	"flag"
	"os"
	"time"
)

const default_interval = 5 * time.Second
const default_rpi = 5
const default_report_interval = 30 * time.Second
const default_runtime = 60 * time.Second

type C struct {
	Interval        time.Duration
	Report_interval time.Duration
	Requests        int64
	Target          string
	Runtime         time.Duration
}

//New creates a new Config from command line arguments.
func New() (c C) {
	cli_interval := flag.Duration("interval", default_interval, " Sets the request interval.")
	cli_report_interval := flag.Duration("report_interval", default_report_interval, " Sets the report display interval.")
	cli_rpi := flag.Int64("rpi", default_rpi, " Sets the number of requests per interval.")
	cli_target := flag.String("url", "", " Sets the target URL.")
	cli_runtime := flag.Duration("runtime", default_runtime, " Sets the amount of time to run the load test for.")
	flag.Parse()
	c.Interval = *cli_interval
	c.Requests = *cli_rpi
	c.Target = *cli_target
	c.Report_interval = *cli_report_interval
	c.Runtime = *cli_runtime
	if c.Target == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	return
}
