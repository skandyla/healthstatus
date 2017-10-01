// Util to parallel ping\loadtest http endpoints
// Useful for testing High Availability of servers and loadbalancers
package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/BurntSushi/toml"
	vegeta "github.com/tsenart/vegeta/lib"
)

type tomlConfig struct {
	Title   string
	Servers map[string]server
}
type server struct {
	InfoUrl   string
	StatusUrl string
	Url       string
	Rate      uint64
	Duration  time.Duration
}

var (
	f      string
	config tomlConfig
)

func main() {
	flag.StringVar(&f, "f", "config.toml", "config file to process")
	flag.Parse()

	//if _, err := toml.DecodeFile(c, &config); err != nil {
	_, err := toml.DecodeFile(f, &config)
	if err != nil {
		log.Fatalf("Error: %v", err)
	}
	for serverName, server := range config.Servers {
		fmt.Printf("Server: %s Endpoint: %s, Rate(rps): %d  Duration(sec): %d\n", serverName, server.Url, server.Rate, server.Duration)
	}
	fmt.Println("")

	start := time.Now()
	ch := make(chan string)
	for _, server := range config.Servers {
		go attack(server.Url, server.Rate, server.Duration, ch) // start a goroutine
	}
	for range config.Servers {
		fmt.Println(<-ch) // receive from channel ch
	}
	fmt.Printf("%.2fs elapsed\n", time.Since(start).Seconds())
}

func attack(url string, rate uint64, duration time.Duration, ch chan<- string) {
	//fmt.Printf("Starting to attack %s\n", url)
	duration = duration * time.Second
	targeter := vegeta.NewStaticTargeter(vegeta.Target{
		Method: "GET",
		URL:    url,
	})
	attacker := vegeta.NewAttacker()

	var metrics vegeta.Metrics
	for res := range attacker.Attack(targeter, rate, duration) {
		metrics.Add(res)
	}
	metrics.Close()

	fmt.Printf("### Endpoint: %s\n", url)
	fmt.Printf("Latencies 99p: %s\n", metrics.Latencies.P99)
	fmt.Printf("Latencies Max: %s\n", metrics.Latencies.Max)
	fmt.Printf("Success Ratio: \t%.2f%%\n", metrics.Success*100)
	fmt.Printf("Status Codes [code:count]: ")
	for code, count := range metrics.StatusCodes {
		fmt.Printf("%s:%d  ", code, count)
	}
	fmt.Println("")
	//r := vegeta.NewTextReporter(&metrics)
	//r.Report(os.Stdout)
	ch <- ""
}
