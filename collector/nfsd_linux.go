// Copyright 2018 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs"
	"github.com/prometheus/procfs/nfs"
)

// A nfsdCollector is a Collector which gathers metrics from /proc/net/rpc/nfsd.
// See: https://www.svennd.be/nfsd-stats-explained-procnetrpcnfsd/
type nfsdCollector struct {
	fs procfs.FS
}

func init() {
	registerCollector("nfsd", defaultEnabled, NewNFSdCollector)
}

const (
	nfsdSubsystem = "nfsd"
)

// NewNFSdCollector returns a new Collector exposing /proc/net/rpc/nfsd statistics.
func NewNFSdCollector() (Collector, error) {
	fs, err := procfs.NewFS(*procPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open procfs: %v", err)
	}

	return &nfsdCollector{
		fs: fs,
	}, nil
}

// Update implements Collector.
func (c *nfsdCollector) Update(ch chan<- prometheus.Metric) error {
	stats, err := c.fs.NFSdServerRPCStats()
	if err != nil {
		return fmt.Errorf("failed to retrieve nfsd stats: %v", err)
	}

	c.updateNFSdReplyCacheStats(ch, &stats.ReplyCache)
	c.updateNFSdFileHandlesStats(ch, &stats.FileHandles)
	c.updateNFSdInputOutputStats(ch, &stats.InputOutput)
	c.updateNFSdThreadsStats(ch, &stats.Threads)
	c.updateNFSdReadAheadCacheStats(ch, &stats.ReadAheadCache)
	c.updateNFSdNetworkStats(ch, &stats.Network)

	return nil
}

// updateNFSdReplyCacheStats collects statistics for the reply cache.
func (c *nfsdCollector) updateNFSdReplyCacheStats(ch chan<- prometheus.Metric, s *nfs.ReplyCache) {
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, nfsdSubsystem, "reply_cache_hits_total"),
			"NFSd Reply Cache client did not receive a reply and decided to re-transmit its request and the reply was cached. (bad).",
			nil,
			nil,
		),
		prometheus.CounterValue,
		float64(s.Hits))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, nfsdSubsystem, "reply_cache_misses_total"),
			"NFSd Reply Cache an operation that requires caching (idempotent).",
			nil,
			nil,
		),
		prometheus.CounterValue,
		float64(s.Misses))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, nfsdSubsystem, "reply_cache_nocache_total"),
			"NFSd Reply Cache non-idempotent operations (rename/delete/…).",
			nil,
			nil,
		),
		prometheus.CounterValue,
		float64(s.NoCache))
}

// updateNFSdFileHandlesStats collects statistics for the file handles.
func (c *nfsdCollector) updateNFSdFileHandlesStats(ch chan<- prometheus.Metric, s *nfs.FileHandles) {
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, nfsdSubsystem, "file_handles_stale_total"),
			"NFSd stale file handles",
			nil,
			nil,
		),
		prometheus.CounterValue,
		float64(s.Stale))
	// NOTE: Other FileHandles entries are unused in the kernel.
}

// updateNFSdInputOutputStats collects statistics for the bytes in/out.
func (c *nfsdCollector) updateNFSdInputOutputStats(ch chan<- prometheus.Metric, s *nfs.InputOutput) {
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, nfsdSubsystem, "disk_bytes_read_total"),
			"NFSd bytes read",
			nil,
			nil,
		),
		prometheus.CounterValue,
		float64(s.Read))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, nfsdSubsystem, "disk_bytes_written_total"),
			"NFSd bytes written",
			nil,
			nil,
		),
		prometheus.CounterValue,
		float64(s.Write))
}

// updateNFSdThreadsStats collects statistics for kernel server threads.
func (c *nfsdCollector) updateNFSdThreadsStats(ch chan<- prometheus.Metric, s *nfs.Threads) {
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, nfsdSubsystem, "server_threads"),
			"NFSd how many kernel threads are running",
			nil,
			nil,
		),
		prometheus.GaugeValue,
		float64(s.Threads))
}

// updateNFSdReadAheadCacheStats collects statistics for the read ahead cache.
func (c *nfsdCollector) updateNFSdReadAheadCacheStats(ch chan<- prometheus.Metric, s *nfs.ReadAheadCache) {
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, nfsdSubsystem, "read_ahead_cache_size_blocks"),
			"NFSd how large the read ahead cache in blocks",
			nil,
			nil,
		),
		prometheus.GaugeValue,
		float64(s.CacheSize))
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, nfsdSubsystem, "read_ahead_cache_not_found_total"),
			"NFSd how large the read ahead cache in blocks",
			nil,
			nil,
		),
		prometheus.CounterValue,
		float64(s.NotFound))
}

// updateNFSdNetworkStats collects statistics for network packets/connections.
func (c *nfsdCollector) updateNFSdNetworkStats(ch chan<- prometheus.Metric, s *nfs.Network) {
	packetDesc := prometheus.NewDesc(
		prometheus.BuildFQName(namespace, nfsdSubsystem, "packets_total"),
		"NFSd how many network packets have been sent/recieved",
		[]string{"proto"},
		nil,
	)
	ch <- prometheus.MustNewConstMetric(
		packetDesc,
		prometheus.CounterValue,
		float64(s.UDPCount), "udp")
	ch <- prometheus.MustNewConstMetric(
		packetDesc,
		prometheus.CounterValue,
		float64(s.TCPCount), "tcp")
	ch <- prometheus.MustNewConstMetric(
		prometheus.NewDesc(
			prometheus.BuildFQName(namespace, nfsdSubsystem, "connections_total"),
			"NFSd how many TCP connections have been made",
			nil,
			nil,
		),
		prometheus.CounterValue,
		float64(s.TCPConnect))
}
