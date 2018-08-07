# rexplorer

rexplorer is a small explorer binary, that can aid in explorering a [tfchain][tfchain] network.
It applies/reverts data —received from an embedded consensus module— into a redis db of choice,
such that the tfchain network data can be consumed/used in a meaningful way.

## Install

```
$ go get -u github.com/threefoldfoundation/rexplorer && rexplorer version
Tool version            v0.1.0
TFChain Daemon version  v1.0.7
Rivine protocol version v1.0.7

Go Version   v1.10.3
GOOS         darwin
GOARCH       amd64
```

## Usage

To start a rexplorer instance for the standard network,
storing all persistent non-redis data into a subdir of the current root directory,
you can do so as simple as:

```
$ rexplorer
2018/08/07 23:58:48 starting rexplorer v0.1.0...
2018/08/07 23:58:48 loading rivine gateway module (1/3)...
2018/08/07 23:58:48 loading rivine consensus module (2/3)...
2018/08/07 23:58:48 loading internal explorer module (3/3)...
```

The persistent dir (used for some local boltdb consensus/gateway data) can be changed
using the `-d`/`--persistent-directory` flag.

Should you want to explore `testnet` instead of the `standard` net you can use the `--network testnet` flag.

For more information use the `--help` flag:

```
$ rexplorer --help
start the rexplorer daemon
Usage:
  rexplorer [flags]
  rexplorer [command]
Available Commands:
  help        Help about any command
  version     show versions of this tool
Flags:
  -h, --help                          help for rexplorer
  -n, --network string                the name of the network to which the daemon connects, one of {standard,testnet} (default "standard")
  -d, --persistent-directory string   location of the root diretory used to store persistent data of the daemon of tfchain
      --redis-addr string             which (tcp) address the redis server listens on (default ":6379")
      --redis-db int                  which redis database slot to use
      --rpc-addr string               which port the gateway listens on (default ":23112")
Use "rexplorer [command] --help" for more information about a command.
```

## Reserved Redis Keys

Ideally you use a Redis database (slot) just for the `rexplorer` instance.
However should you not be able to allocate an entire database (slot) just for the `rexplorer instance`,
please do not ever touch the reserved keys. You'll break your own explored data should you write/delete any values
stored directly or indirectly of a reserved key.

There are two types of keys:

* internal keys: these are keys which are meant for internals of the `rexplorer` instance, and are not meant for public consumption;
* public keys: these are keys meant for public consumption and have a well-defined format;

Following _internal_ keys are reserved:

* `<chainName>:<networkName>:state`:
    * used for internal state of this explorer, in JSON format
    * format value: JSON
    * example key: `tfchain:standard:state`
* `<chainName>:<networkName>:ucos`:
    * all unspent coin outputs, and for each coin output only the info which is required for the inner workings of the `rexplorer`
    * format value: custom
    * example key: `tfchain:testnet:ucos`
* `<chainName>:<networkName>:lcos.height:<height>`:
    * all locked coin outputs on a given height
    * format value: custom
    * example key: `tfchain:standard:lcos.height:42`
* `<chainName>:<networkName>:lcos.time:<timestamp[:-5]>`:
    * all locked coin outputs for a given timestmap range
    * format value: custom
    * example key: `tfchain:standard:lcos.time:15341`

Following _public_ keys are reserved:

* `<chainName>:<networkName>:stats`:
    * used for global network statistics
    * format value: JSON
    * example key: `tfchain:standard:stats`
* `<chainName>:<networkName>:addresses`:
    * set of unique wallet addresses used (even if reverted) in the network
    * format value: [Redis SET][redistypes], where each value is a [Rivine][rivine]-defined hex-encoded UnlockHash
    * example key: `tfchain:standard:addresses`
* `<chainName>:<networkName>:address:<unlockHashHex>:balance`:
    * used by all wallet addresses, contains both locked and unlocked (coin) balance
    * format value: JSON
    * example key: `tfchain:testnet:address:0178f4ea48f511d1a59f90bd44f237c7b2e7016557ce74eb688419f53764a91543b4466b2ff481:balance`
* `<chainName>:<networkName>:address:<unlockHashHex>:outputs.locked`:
    * used to store locked (by time or blockHeight) outputs destined for an address
    * format value: [Redis HASHMAP][redistypes], where each key is hex-encoded CoinOutputID and the value being the [Rivine][rivine]-defined JSON-encoded CoinOutput
    * example key: `tfchain:standard:address:0104089348845d5465887affb55f2133b8cdee789ddfd1b0c2f3400f2a41d1a547ecc41f029c50:outputs.locked`
* `<chainName>:<networkName>:address:<unlockHashHex>:multisig.addresses`:
    * used in both directions for multisig (wallet) addresses (see [the Get MultiSig Addresses example](#get-multisig-addresses) for more information)
    * format value: [Redis SET][redistypes], where each value is a [Rivine][rivine]-defined hex-encoded UnlockHash
    * example key: `tfchain:testnet:address:01b650391f06c6292ecf892419dd059c6407bf8bb7220ac2e2a2df92e948fae9980a451ac0a6aa:multisig.addresses`

Rivine Value Encodings:

* addresses are Hex-encoded and the exact format (and how it is created) is described in:
  <https://github.com/rivine/rivine/blob/master/doc/transactions/unlockhash.md#textstring-encoding>
* currencies are encoded as described in <https://godoc.org/math/big#Int.Text>
  using base 10, and using the smallest coin unit as value (e.g. 10^-9 TFT)
* coin outputs are stored in the Rivine-defined JSON format, described in:
  <https://github.com/rivine/rivine/blob/master/doc/transactions/transaction.md#json-encoding-of-outputs-in-v0-transactions> (`v0` tx) and
  <https://github.com/rivine/rivine/blob/master/doc/transactions/transaction.md#json-encoding-of-outputs-in-v1-transactions> (`v1` tx)

JSON formats of value types defined by this module:

* example of global stats (stored under `<chainName>:<networkName>:stats`):

```json
{
    "timestamp": 1533670089,
    "blockHeight": 76824,
    "txCount": 77139,
    "valueTxCount": 316,
    "coinOutputCount": 1211,
    "coinInputCount": 355,
    "minerPayoutCount": 77062,
    "minerPayouts": "76855400000001",
    "coins": "695175855400000001"
}
```
* example of wallet balance (stored under `<chainName>:<networkName>:address:<unlockHashHex>:balance`):

```json
{
    "locked": "0",
    "unlocked": "250000000000"
}
```

## Examples

These examples assume you have a `rexplorer` instance running (and synced!!!),
using the default redis address (`:6379`) and default db slot (`0`).

### Get Coins

There is a Go example that you can checkout at [/examples/getcoins/main.go](/examples/getcoins/main.go),
and you can run it yourself as follows:

```
$ go run ./examples/getcoins/main.go -network=testnet 0178f4ea48f511d1a59f90bd44f237c7b2e7016557ce74eb688419f53764a91543b4466b2ff481
unlocked: 42.3 TFT
locked: 0 TFT
```

You can run the same example directly from the shell —using `redis-cli`— as well:

```
$ redis-cli get tfchain:testnet:address:0178f4ea48f511d1a59f90bd44f237c7b2e7016557ce74eb688419f53764a91543b4466b2ff481:balance
"{\"locked\":\"0\",\"unlocked\":\"42300000000\"}"
```

As you can see for yourself, the balance of an address is stored as a JSON object.

### Get All Unique Addresses Used

Get all the unique addresses used within a network.
Even if an address is only used in a reverted block, it is still tracked and kept:

```
$ redis-cli smembers tfchain:standard:addresses
1) 01fea3ae2854f6e497c92a1cdd603a0bc92ada717200e74f64731e86a923479883519804b18d9d
2) 01fef1037d0e51042838e4265a1af4f753b8f69de5a7be85a5f3a3c6bd1fbcb8f20986b4aae3a5
3) 0148d275cffe21a79a865d78529682e347d56615e0033ff114731014349b970c033acae5fbf3a3
4) 01cc1872da1c5b2f6bc02fead5f660992477b7c3d7133c75746b7adeec72bdda5c9149cb36e34a
5) 01cc6173a28b18ce172466c2b7aca93465ff8c2ccebbad27c4c54bada8a80e8de667c0a0ae0e5f
...
630) 017c69af16f91cfe3360e39205411b15f9a0f6cb7502e2d4cbf7c428d44595b9f3a4b377740bfe
631) 01806f23a376c216ca96a2dc0b65f74ad47bcd13ae4d65b8af65211fa6540cc7ccd270c647d443
632) 0142c3d75a2ce6052e316e0d61f290cdc2e974de9107e26185703722ab2c6c0f6d203d0f8341ca
633) 01060d90351de9bf9892711c713e42622cf4c8743e08e5ee800da4a393446ded584cb8bf8250d8
634) 018f199d998e248936eb4ac7a78f6084b6613d7611b7b0f2838dd245c4217ff2413b10d54d070f
635) 01fdd6c673686ccf1aa2caa5c0ea08bcb25e8fe3cbbc1079b350500314ef8defa4991cdf27bc1f
```

If you pipe the command from this example into the command of [the Get Coins example](#get-coins),
you'll be able to get the balance of each of the wallets in existence of a network.

Following this example we can see how to get the amount of unique addresses used in a network:

```
$ redis-cli scard tfchain:standard:addresses
(integer) 635
```

### Get Global Statistics

There is a Go example that you can checkout at [/examples/getstats/main.go](/examples/getstats/main.go),
and you can run it yourself as follows:

```
$ go run ./examples/getstats/main.go
tfchain/testnet has:
  * a total of 100559305.60000007 TFT, of which 559305.60000007 TFT are payed out as fees and miner rewards
  * a block height of 55929, with the time of the highest block being 2018-06-26 16:30:19 +0200 CEST (1530023419)
  * a total of 55930 blocks, 163 value transactions, 311 coin outputs, 56062 miner payouts and 1465 coin inputs
  * a total of 275 unique wallet addresses that have been used
  * an average of 1.907975 coin outputs per value transaction
  * an average of 0.002914 value transactions per block
  * 0.290604% value transactions of a total of 56090 transactions
```

You can run the same example directly from the shell —using `redis-cli`— as well:

```
$ redis-cli get tfchain:testnet:stats
"{\"timestamp\":1530023419,\"blockHeight\":55929,\"txCount\":56090,\"valueTxCount\":163,\"coinOutputCount\":311,\"coinInputCount\":1465,\"minerPayoutCount\":56062,\"minerPayouts\":\"559305600000070\",\"coins\":\"100559305600000070\"}"
```

As you can see for yourself, the balance of an address is stored as a JSON object.
In the Golang example we added some extra logic to showcase some examples of
some statistics you can compute based on the tracked global statistical values.

### Get MultiSig Addresses

There is a Go example that you can checkout at [/examples/getmultisigaddresses/main.go](/examples/getmultisigaddresses/main.go),
and you can run it yourself as follows:

```
$ go run ./examples/getmultisigaddresses/main.go -network=testnet 01b650391f06c6292ecf892419dd059c6407bf8bb7220ac2e2a2df92e948fae9980a451ac0a6aa
* 0359aaaa311a10efd7762953418b828bfe2d4e2111dfe6aaf82d4adf6f2fb385688d7f86510d37
```

You can run the same example directly from the shell —using `redis-cli`— as well:

```
$ redis-cli smembers tfchain:testnet:address:01b650391f06c6292ecf892419dd059c6407bf8bb7220ac2e2a2df92e948fae9980a451ac0a6aa:multisig.addresses
1) "0359aaaa311a10efd7762953418b828bfe2d4e2111dfe6aaf82d4adf6f2fb385688d7f86510d37"
```

This example also works in the opposite direction, where the multisig address will return all owner addresses:

```
$ redis-cli smembers tfchain:testnet:address:0359aaaa311a10efd7762953418b828bfe2d4e2111dfe6aaf82d4adf6f2fb385688d7f86510d37:multisig.addresses
1) "01b650391f06c6292ecf892419dd059c6407bf8bb7220ac2e2a2df92e948fae9980a451ac0a6aa"
2) "0114df42a3bb8303a745d23c47062a1333246b3adac446e6d62f4de74f5223faf4c2da465e76af"
```

### Get Balance of all MultiSig Owners

Should we want to know who is the richest owner of a MultiSig wallet, we can do so by combinding some of the dumped data:

```
$ redis-cli smembers tfchain:testnet:address:0359aaaa311a10efd7762953418b828bfe2d4e2111dfe6aaf82d4adf6f2fb385688d7f86510d37:multisig.addresses | xargs -I % sh -c 'echo %; redis-cli get tfchain:testnet:address:%:balance'
01b650391f06c6292ecf892419dd059c6407bf8bb7220ac2e2a2df92e948fae9980a451ac0a6aa
(nil)
0114df42a3bb8303a745d23c47062a1333246b3adac446e6d62f4de74f5223faf4c2da465e76af
"{\"locked\":\"0\",\"unlocked\":\"0\"}"
```

A `nil` balance JSON object should be accepted as a balance object with all currency properties equal to `0`.

[tfchain]: https://github.com/threefoldfoundation/tfchain
[rivine]: https://github.com/rivine/rivine
[redistypes]: https://redis.io/topics/data-types