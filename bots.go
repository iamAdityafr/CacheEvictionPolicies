package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// doing mock symbols
var symbols = []string{
	"BTCUSDT", "ETHUSDT", "BNBUSDT", "ADAUSDT", "XRPUSDT",
	"DOGEUSDT", "SOLUSDT", "DOTUSDT", "LTCUSDT", "LINKUSDT",
	"TRXUSDT", "MATICUSDT", "SHIBUSDT", "ATOMUSDT", "ETCUSDT",
	"XMRUSDT", "AVAXUSDT", "NEARUSDT", "BCHUSDT", "ALGOUSDT",
	"VETUSDT", "FILUSDT", "ICPUSDT", "XTZUSDT", "EOSUSDT",
	"THETAUSDT", "AAVEUSDT", "KSMUSDT", "MKRUSDT", "COMPUSDT",
	"SUSHIUSDT", "YFIUSDT", "1INCHUSDT", "CAKEUSDT", "FTMUSDT",
	"RUNEUSDT", "ZILUSDT", "HNTUSDT", "CHZUSDT", "ENJUSDT",
	"CELOUSDT", "BATUSDT", "KNCUSDT", "DGBUSDT", "ONTUSDT",
	"QTUMUSDT", "XEMUSDT", "SCUSDT", "HOTUSDT", "RVNUSDT",
	"ARUSDT", "EGLDUSDT", "STXUSDT", "SANDUSDT", "GRTUSDT",
	"NEOUSDT", "WAVESUSDT", "KAVAUSDT", "CVCUSDT", "ANKRUSDT",
	"ZRXUSDT", "OMGUSDT", "OCEANUSDT", "DCRUSDT", "BTGUSDT",
	"LRCUSDT", "ICXUSDT", "BNTUSDT", "HBARUSDT", "CELRUSDT",
	"ORNUSDT", "REEFUSDT", "RSRUSDT", "FETUSDT", "NKNUSDT",
	"STORJUSDT", "IOSTUSDT", "MDXUSDT", "LPTUSDT", "TWTUSDT",
	"ANKRUSDT", "KLAYUSDT", "KSMUSDT", "CHSBUSDT", "SUPERUSDT",
	"VGXUSDT", "REEFUSDT", "CTSIUSDT", "ALICEUSDT", "BTSUSDT",
	"FXSUSDT", "CVXUSDT", "ENSUSDT", "FLMUSDT", "GLMRUSDT",
	"ACHUSDT", "COTIUSDT", "RLCUSDT", "TRIBEUSDT", "SPELLUSDT",
	"FXSUSDT", "GMXUSDT", "OPUSDT", "PEOPLEUSDT", "AGIXUSDT",
	"RNDRUSDT", "DYDXUSDT", "ARPAUSDT", "LDOUSDT", "SUSHIUSDT",
}

const botsNum = 20

const reqIvl = 200 * time.Millisecond

func main() {
	rand.Seed(time.Now().UnixNano())
	var wg sync.WaitGroup

	for i := 0; i < botsNum; i++ {
		wg.Add(1)
		go func(botId int) {
			defer wg.Done()
			for j := 0; j < 50; j++ { // each bot hits 50 requests
				symbol := symbols[rand.Intn(len(symbols))]
				hitEndpoint(symbol, botId)
				time.Sleep(reqIvl + time.Duration(rand.Intn(200))*time.Millisecond)
			}
		}(i + 1)
	}

	wg.Wait()
	fmt.Println("all bots job done.")
}

func hitEndpoint(symbol string, botId int) {
	url := fmt.Sprintf("http://localhost:42069/quote?symbol=%s", symbol)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Printf("Bot %d:  Error fetching %s: %v\n", botId, symbol, err)
		return
	}
	defer resp.Body.Close()
	fmt.Printf("Bot %d:  Fetched %s - With Status: %d\n", botId, symbol, resp.StatusCode)
}
