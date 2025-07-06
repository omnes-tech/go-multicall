# `multicall`: Perform multiple calls using only one RPC call

[![Go Reference](https://pkg.go.dev/badge/github.com/omnes-tech/multicall.svg)](https://pkg.go.dev/github.com/omnes-tech/multicall)
[![Go Report Card](https://goreportcard.com/badge/github.com/omnes-tech/multicall)](https://goreportcard.com/report/github.com/omnes-tech/multicall)
[![Coverage Status](https://coveralls.io/repos/github/omnes-tech/multicall/badge.svg?branch=main)](https://coveralls.io/github/omnes-tech/multicall?branch=main)
[![Latest Release](https://img.shields.io/github/v/release/omnes-tech/multicall)](https://github.com/omnes-tech/multicall/releases/latest)
<!-- <img src="https://w3.cool/gopher.png" align="right" alt="W3 Gopher" width="158" height="224" -->

Carry out several calls with only one RPC call.

```shell
go get github.com/omnes-tech/multicall
```

## At a Glance

Instantiate the multicall client:
```go
client, err := multicall.NewClient("http://localhost:8545", nil)
if err!= nil {
    log.Fatal(err)
}
```

Now you just need to call any method you need!

Write (transaction) functions:
- `AggregateCalls`
- `TryAggregateCalls`
- `TryAggregateCalls3`

Read (call) functions:
- `SimulateCall`
- `AggregateStatic`
- `TryAggregateStatic`
- `TryAggregateStatic3`
- `CodeLengths`
- `Balances`
- `AddressesData`

## Deployed Smart Contracts

Check out the deployed addresses [here](https://github.com/omnes-tech/multicall-contract/blob/main/README.md#deployments) on different chains.