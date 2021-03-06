package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/rivine/rivine/types"

	"github.com/gomodule/redigo/redis"
)

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		panic("usage: " + os.Args[0] + " <unlockhash>")
	}
	var uh types.UnlockHash
	err := uh.LoadString(args[0])
	if err != nil {
		fmt.Fprintln(os.Stderr, "usage: "+os.Args[0]+" <unlockhash>")
		panic(fmt.Sprintf("invalid uh %q: %v", args[0], err))
	}

	conn, err := redis.Dial("tcp", dbAddress, redis.DialDatabase(dbSlot))
	if err != nil {
		panic(err)
	}

	addressKey, addressField := getAddressKeyAndField(uh)
	var wallet struct {
		MultiSignAddresses []types.UnlockHash `json:"multisignAddresses,omitempty"`
	}
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

	// print all unlock hashes
	for _, uh := range wallet.MultiSignAddresses {
		fmt.Println("* " + uh.String())
	}
}

func getAddressKeyAndField(uh types.UnlockHash) (key, field string) {
	str := uh.String()
	key, field = "a:"+str[:6], str[6:]
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
