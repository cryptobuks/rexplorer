package main

import (
	"encoding/json"
	"flag"
	"fmt"

	"github.com/rivine/rivine/types"

	"github.com/gomodule/redigo/redis"
)

func main() {
	flag.Parse()

	conn, err := redis.Dial("tcp", dbAddress, redis.DialDatabase(dbSlot))
	if err != nil {
		panic(err)
	}
	// get stats, so we know what are the to be expected total coins and total locked coins
	b, err := redis.Bytes(conn.Do("GET", "stats"))
	if err != nil {
		panic("failed to get network stats: " + err.Error())
	}
	var stats struct {
		BlockHeight types.BlockHeight `json:"blockHeight"`
		Coins       types.Currency    `json:"coins"`
		LockedCoins types.Currency    `json:"lockedCoins"`
	}
	err = json.Unmarshal(b, &stats)
	if err != nil {
		panic("failed to json-unmarshal network stats: " + err.Error())
	}

	// get all unique addresses
	addresses, err := redis.Strings(conn.Do("SMEMBERS", "addresses"))
	if err != nil {
		panic("failed to get all unique addresses: " + err.Error())
	}

	// compute total unlocked and locked coins for all addresses
	var unlockedCoins, lockedCoins types.Currency
	for _, addr := range addresses {
		var wallet struct {
			Balance struct {
				Unlocked types.Currency `json:"unlocked"`
				Locked   struct {
					Total types.Currency `json:"total"`
				} `json:"locked"`
			} `json:"balance,omitempty"`
		}
		addressKey, addressField := getAddressKeyAndField(addr)
		b, err := redis.Bytes(conn.Do("HGET", addressKey, addressField))
		if err != nil {
			if err != redis.ErrNil {
				panic("failed to get wallet " + err.Error())
			}
			b = []byte("{}")
		}
		err = json.Unmarshal(b, &wallet)
		if err != nil {
			panic("failed to json-unmarshal wallet: " + err.Error())
		}
		unlockedCoins = unlockedCoins.Add(wallet.Balance.Unlocked)
		lockedCoins = lockedCoins.Add(wallet.Balance.Locked.Total)
	}
	totalCoins := unlockedCoins.Add(lockedCoins)

	// ensure our total coin count is as expected
	if c := lockedCoins.Cmp(stats.LockedCoins); c != 0 {
		var diff types.Currency
		switch c {
		case -1:
			diff = stats.LockedCoins.Sub(lockedCoins)
		case 1:
			diff = lockedCoins.Sub(stats.LockedCoins)
		}

		panic(fmt.Sprintf("unexpected locked coins: %s != %s (diff: %s)",
			lockedCoins.String(), stats.LockedCoins.String(), diff.String()))
	}
	if c := totalCoins.Cmp(stats.Coins); c != 0 {
		var diff types.Currency
		switch c {
		case -1:
			diff = stats.Coins.Sub(totalCoins)
		case 1:
			diff = totalCoins.Sub(stats.Coins)
		}

		panic(fmt.Sprintf("unexpected total coins: %s != %s (diff: %s)",
			totalCoins.String(), stats.Coins.String(), diff.String()))
	}

	fmt.Printf(
		"sumcoins test on block height %d passed :)\n", stats.BlockHeight)
}

func getAddressKeyAndField(addr string) (key, field string) {
	key, field = "a:"+addr[:6], addr[6:]
	return
}

var (
	dbAddress string
	dbSlot    int
)

func init() {
	flag.StringVar(&dbAddress, "db-address", ":6379", "(tcp) address of the redis db")
	flag.IntVar(&dbSlot, "db-slot", 0, "slot/index of the redis db")
}
