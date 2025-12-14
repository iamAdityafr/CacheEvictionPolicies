package main

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"sort"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// the value to be cached
type Ticker struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price"`
	Time   string  `json:"time"`
}

var (
	cache       *LFUCache[string, Ticker]
	cacheHits   int64
	cacheMisses int64

	cacheHitsGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cache_hits_total",
		Help: "Total number of cache hits.",
	})
	cacheMissesGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cache_misses_total",
		Help: "Total number of cache misses.",
	})
	cacheSizeGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cache_size",
		Help: "Current items number in cache.",
	})
	cacheHitRatioGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "cache_hit_ratio",
		Help: "Ratio of cache hits.",
	})
	reqDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "request_duration_seconds",
		Help:    "request latency",
		Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
	}, []string{"symbol", "cache"})
	topSymbolsGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "top_symbols_requests",
			Help: "freq of top 5 requested symbols.",
		},
		[]string{"symbol"},
	)
)

func init() {
	prometheus.MustRegister(cacheHitsGauge, cacheMissesGauge, cacheSizeGauge, cacheHitRatioGauge, reqDuration, topSymbolsGauge)
}
func main() {

	var err error
	cache, err = NewLFUCache[string, Ticker](10)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/quote", cstmHandler)
	http.Handle("/metrics", promhttp.Handler())

	go dometrics()

	fmt.Println("Running at :42069 ...")
	if err := http.ListenAndServe(":42069", nil); err != nil {
		panic(err)
	}
}

func cstmHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	symbol := r.URL.Query().Get("symbol")
	if symbol == "" {
		http.Error(w, "symbol not found", http.StatusBadRequest)
		return
	}

	var cacheHit string
	var t Ticker

	if ticker, ok := cache.Get(symbol); ok {
		atomic.AddInt64(&cacheHits, 1)
		cacheHit = "hit"
		t = ticker
	} else {
		atomic.AddInt64(&cacheMisses, 1)
		cacheHit = "miss"
		t = Ticker{
			Symbol: symbol,
			Price:  1000 + rand.Float64()*50000,
			Time:   time.Now().Format(time.RFC3339),
		}
		cache.Put(symbol, t)
	}
	duration := time.Since(start).Seconds()
	reqDuration.WithLabelValues(symbol, cacheHit).Observe(duration)
	cache.Put(symbol, t)
	json.NewEncoder(w).Encode(t)
}

func dometrics() {
	for {
		hits := atomic.LoadInt64(&cacheHits)
		misses := atomic.LoadInt64(&cacheMisses)
		sz := len(cache.Cache)

		cacheHitsGauge.Set(float64(hits))
		cacheMissesGauge.Set(float64(misses))
		cacheSizeGauge.Set(float64(sz))

		total := float64(hits + misses)
		if total > 0 {
			cacheHitRatioGauge.Set(float64(hits) / total)
		} else {
			cacheHitRatioGauge.Set(0)
		}

		updateTopRequests()

		time.Sleep(1 * time.Second)
	}
}

func updateTopRequests() {
	topSymbolsGauge.Reset()
	items := cache.Items()

	type topFreqVal struct {
		Key  string
		Freq int
	}
	var topFreqList []topFreqVal
	for k, v := range items {
		topFreqList = append(topFreqList, topFreqVal{k, v.Freq})
	}

	sort.Slice(topFreqList, func(i, j int) bool {
		return topFreqList[i].Freq > topFreqList[j].Freq
	})

	for i := 0; i < len(topFreqList) && i < 5; i++ {
		topSymbolsGauge.WithLabelValues(topFreqList[i].Key).Set(float64(topFreqList[i].Freq))
	}
}
