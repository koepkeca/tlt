package config

import (
	"flag"
	"log"
	"net/http"
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
	Target          []string
	Runtime         time.Duration
	Outfile         string
}

// New creates a new Config from command line arguments.
func New() (c C) {
	cli_interval := flag.Duration("interval", default_interval, " Sets the request interval.")
	cli_report_interval := flag.Duration("report_interval", default_report_interval, " Sets the report display interval.")
	cli_rpi := flag.Int64("rpi", default_rpi, " Sets the number of requests per interval.")
	cli_target := flag.String("url", "", " Sets the target URL. [Cannot be used with xhr replay mode]")
	cli_runtime := flag.Duration("runtime", default_runtime, " Sets the amount of time to run the load test for.")
	cli_xhrfile := flag.String("xhr", "", "Sets the source xhr file and activates XHR Replay mode. [Cannot be used with url.]")
	cli_outfile := flag.String("outfile", "session.json", "Sets the output file")
	flag.Parse()
	if *cli_target != "" && *cli_xhrfile != "" {
		log.Printf("The url and xhr paramaters may not be used together. Please pick one mode and try again.")
		flag.PrintDefaults()
		os.Exit(1)
	}
	if *cli_target == "" && *cli_xhrfile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
	c.Target = []string{*cli_target}
	if *cli_xhrfile != "" {
		fp, e := os.Open(*cli_xhrfile)
		if e != nil {
			log.Printf("XHR file not found or unable to open. Quitting")
			flag.PrintDefaults()
			os.Exit(1)
		}
		xhr, e := NewXhrFromReader(fp)
		if e != nil {
			log.Printf("Error reading xhr file %s", e)
			os.Exit(1)
		}
		c.Target = []string{}
		for _, nextTgt := range xhr.Log.Entries {
			if nextTgt.Response.Status != http.StatusOK {
				log.Printf("Skipping %s, imported response code is: %s", nextTgt.Request.URL, nextTgt.Response.Status)
				continue
			}
			log.Printf("%s\n", nextTgt.Request.URL)
			c.Target = append(c.Target, nextTgt.Request.URL)
		}
	}
	c.Interval = *cli_interval
	c.Requests = *cli_rpi
	c.Report_interval = *cli_report_interval
	c.Runtime = *cli_runtime
	c.Outfile = *cli_outfile
	return
}
