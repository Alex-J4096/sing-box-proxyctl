/*
Copyright © 2026 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Alex-J4096/proxyctl/util"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

var pingConfigPath string
var pingCachePath string
var pingTimeout time.Duration
var pingConcurrency int

type geoLookupResponse struct {
	Status     string `json:"status"`
	Country    string `json:"country"`
	RegionName string `json:"regionName"`
	City       string `json:"city"`
	Query      string `json:"query"`
	Message    string `json:"message"`
}

type pingJob struct {
	Index    int
	Outbound util.Outbound
}

type pingResult struct {
	Job    pingJob
	Result util.NodePingResult
}

func readSingboxConfig(path string) (util.SingboxConfig, error) {
	var cfg util.SingboxConfig

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func tcpPing(outbound util.Outbound, timeout time.Duration) (int64, error) {
	address := net.JoinHostPort(outbound.Server, strconv.Itoa(outbound.ServerPort))
	start := time.Now()

	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	return time.Since(start).Milliseconds(), nil
}

func lookupRegion(server string, timeout time.Duration) (string, string) {
	client := http.Client{Timeout: timeout}
	endpoint := "http://ip-api.com/json/" + url.PathEscape(server) + "?fields=status,country,regionName,city,query,message"

	resp, err := client.Get(endpoint)
	if err != nil {
		return "unknown", ""
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "unknown", ""
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "unknown", ""
	}

	var geo geoLookupResponse
	if err := json.Unmarshal(body, &geo); err != nil || geo.Status != "success" {
		return "unknown", ""
	}

	country := strings.TrimSpace(geo.Country)
	if country == "" {
		return "unknown", geo.Query
	}
	return country, geo.Query
}

func pingNode(job pingJob, timeout time.Duration) util.NodePingResult {
	outbound := job.Outbound
	now := time.Now().Format(time.RFC3339)
	result := util.NodePingResult{
		Tag:        outbound.Tag,
		Server:     outbound.Server,
		ServerPort: outbound.ServerPort,
		Status:     "fail",
		Region:     "unknown",
		CheckedAt:  now,
	}

	var wg sync.WaitGroup
	var latencyErr error
	wg.Add(2)

	go func() {
		defer wg.Done()
		latency, err := tcpPing(outbound, timeout)
		if err != nil {
			latencyErr = err
			return
		}
		result.Status = "ok"
		result.LatencyMS = latency
	}()

	go func() {
		defer wg.Done()
		region, ip := lookupRegion(outbound.Server, timeout)
		result.Region = region
		result.IP = ip
	}()

	wg.Wait()
	if latencyErr != nil {
		result.Error = latencyErr.Error()
	}
	return result
}

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping [id]",
	Short: "Ping nodes concurrently and save results.",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := readSingboxConfig(pingConfigPath)
		if err != nil {
			pterm.Error.Println("failed to read config:", err)
			return
		}
		if len(cfg.Outbounds) == 0 {
			pterm.Warning.Println("no nodes found in config")
			return
		}

		jobs := []pingJob{}
		if len(args) == 1 {
			id, err := strconv.Atoi(args[0])
			if err != nil || id < 0 || id >= len(cfg.Outbounds) {
				pterm.Error.Println("invalid node id:", args[0])
				return
			}
			jobs = append(jobs, pingJob{Index: id, Outbound: cfg.Outbounds[id]})
		} else {
			for i, outbound := range cfg.Outbounds {
				jobs = append(jobs, pingJob{Index: i, Outbound: outbound})
			}
		}

		if pingConcurrency < 1 {
			pingConcurrency = 1
		}
		if pingConcurrency > len(jobs) {
			pingConcurrency = len(jobs)
		}

		cachePath := pingCachePath
		if cachePath == "" {
			cachePath = util.DefaultPingCachePath(pingConfigPath)
		}

		cache, err := util.LoadPingCache(cachePath)
		if err != nil {
			pterm.Warning.Println("failed to read existing ping cache:", err)
		}

		pterm.Info.Println(fmt.Sprintf("pinging %d node(s), concurrency=%d, timeout=%s", len(jobs), pingConcurrency, pingTimeout))

		jobCh := make(chan pingJob)
		resultCh := make(chan pingResult)
		var wg sync.WaitGroup

		for worker := 0; worker < pingConcurrency; worker++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for job := range jobCh {
					result := pingNode(job, pingTimeout)
					resultCh <- pingResult{Job: job, Result: result}
				}
			}()
		}

		go func() {
			for _, job := range jobs {
				jobCh <- job
			}
			close(jobCh)
			wg.Wait()
			close(resultCh)
		}()

		resultsByID := map[int]util.NodePingResult{}

		for done := range resultCh {
			cache.Results[util.PingCacheKey(done.Job.Outbound, done.Job.Index)] = done.Result
			resultsByID[done.Job.Index] = done.Result
		}

		tableData := pterm.TableData{
			{"id", "status", "latency", "region", "tag"},
		}

		for _, job := range jobs {
			result := resultsByID[job.Index]
			status := result.Status
			if result.Error != "" {
				status = "fail"
			}
			tableData = append(tableData, []string{
				strconv.Itoa(job.Index),
				status,
				util.FormatLatency(result),
				util.FormatRegion(result.Region),
				displayTag(result.Tag),
			})
		}

		if err := util.SavePingCache(cachePath, cache); err != nil {
			pterm.Error.Println("failed to save ping cache:", err)
			return
		}

		if err := pterm.DefaultTable.
			WithHasHeader().
			WithHeaderRowSeparator("-").
			WithData(tableData).
			Render(); err != nil {
			pterm.Error.Println("failed to render ping results:", err)
			return
		}

		pterm.Success.Println("ping results saved to " + cachePath)
	},
}

func init() {
	nodeCmd.AddCommand(pingCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// pingCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	pingCmd.Flags().StringVarP(&pingConfigPath, "config", "c", "./config.json", "sing-box config path")
	pingCmd.Flags().StringVar(&pingCachePath, "ping-cache", "", "ping result cache path")
	pingCmd.Flags().DurationVarP(&pingTimeout, "timeout", "t", 3*time.Second, "TCP ping and region lookup timeout")
	pingCmd.Flags().IntVarP(&pingConcurrency, "concurrency", "j", 8, "maximum concurrent ping workers")
}
