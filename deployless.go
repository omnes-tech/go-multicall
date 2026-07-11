package multicall

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/omnes-tech/abi"
)

type CallType uint8

const (
	SIMULATE_CALL = iota
	SIMULATE_DELEGATE_CALL
	STATIC_CALL
	TRY_STATIC_CALL
	TRY_STATIC_CALL2
	CODE_LENGTH
	BALANCES
	ADDRESSES_DATA
	CHAIN_DATA
)

type CallsWithRequireSuccess struct {
	Calls          Calls
	RequireSuccess bool
}

func deploylessSimulation(calls Calls, client *ethclient.Client, from *common.Address, blockNumber *big.Int, overrides StateOverride) Result {
	arrayfiedCalls, _, err := calls.ToArray(true, false)
	if err != nil {
		return Result{Success: false, Error: err}
	}

	_, txOrCall, err := makeDeploylessCall(
		arrayfiedCalls,
		false,
		SIMULATE_CALL,
		from,
		client,
		[]string{"(address,bytes,uint256)[]"},
		blockNumber,
		overrides,
	)
	if err != nil {
		if strings.Contains(err.Error(), "execution reverted") {
			encodedRevert, ok := parseRevertData(err)
			if ok {
				decodedRevert, err := abi.DecodeWithSignature(
					"MultiCall__Simulation((bool,bytes,uint256)[])",
					encodedRevert,
				)
				if err != nil {
					return Result{Success: false, Error: err, TxOrCall: txOrCall}
				}
				decodedRevert = decodedRevert[0].([]any)

				for i, result := range decodedRevert {
					decodedRevert[i].([]any)[1] = common.Bytes2Hex(result.([]any)[1].([]byte))
				}

				return Result{
					Success:  true,
					Result:   decodedRevert,
					TxOrCall: txOrCall,
				}
			}
		}
		return Result{Success: false, Error: err, TxOrCall: txOrCall}
	}

	return Result{Success: false, Error: fmt.Errorf("call did not returned simulation result"), TxOrCall: txOrCall}
}

func deploylessAggregateStatic(calls Calls, client *ethclient.Client, from *common.Address, blockNumber *big.Int, overrides StateOverride) Result {
	arrayfiedCalls, _, err := calls.ToArray(false, false)
	if err != nil {
		return Result{Success: false, Error: err}
	}

	rawResponse, txOrCall, err := makeDeploylessCall(
		arrayfiedCalls,
		false,
		STATIC_CALL,
		from,
		client,
		[]string{"(address,bytes)[]"},
		blockNumber,
		overrides,
	)
	if err != nil {
		return Result{Success: false, Error: err, TxOrCall: txOrCall}
	}

	resultArgs, err := abi.Decode([]string{"bytes[]"}, common.Hex2Bytes(rawResponse[2:]))
	if err != nil {
		return Result{Success: false, Error: err, TxOrCall: txOrCall}
	}
	resultArgs = resultArgs[0].([]any)

	var result []any
	for i, call := range calls {
		result_i, err := abi.Decode(call.ReturnTypes, resultArgs[i].([]byte))
		if err != nil {
			return Result{Success: false, Error: err, TxOrCall: txOrCall}
		}

		result = append(result, result_i)
	}

	return Result{Success: true, Result: result, TxOrCall: txOrCall}
}

func deploylessTryAggregateStatic(
	calls Calls, requireSuccess bool, client *ethclient.Client, from *common.Address, blockNumber *big.Int, overrides StateOverride,
) Result {
	arrayfiedCalls, _, err := calls.ToArray(false, false)
	if err != nil {
		return Result{Success: false, Error: err}
	}

	rawResponse, txOrCall, err := makeDeploylessCall(
		arrayfiedCalls,
		requireSuccess,
		TRY_STATIC_CALL,
		from,
		client,
		[]string{"(address,bytes)[]", "bool"},
		blockNumber,
		overrides,
	)
	if err != nil {
		return Result{Success: false, Error: err, TxOrCall: txOrCall}
	}

	resultArgs, err := abi.Decode([]string{"(bool,bytes)[]"}, common.Hex2Bytes(rawResponse[2:]))
	if err != nil {
		return Result{Success: false, Error: err, TxOrCall: txOrCall}
	}
	resultArgs = resultArgs[0].([]any)

	var result []any
	for i, call := range calls {
		resultArgs[i].([]any)[1], err = abi.Decode(call.ReturnTypes, resultArgs[i].([]any)[1].([]byte))
		if err != nil {
			return Result{Success: false, Error: err, TxOrCall: txOrCall}
		}

		result = append(result, resultArgs[i])
	}

	return Result{Success: true, Result: result, TxOrCall: txOrCall}
}

func deploylessTryAggregateStatic3(
	calls CallsWithFailure, client *ethclient.Client, from *common.Address, blockNumber *big.Int, overrides StateOverride,
) Result {
	arrayfiedCalls, _, err := calls.ToArray(false, false)
	if err != nil {
		return Result{Success: false, Error: err}
	}

	rawResponse, txOrCall, err := makeDeploylessCall(
		arrayfiedCalls,
		false,
		TRY_STATIC_CALL2,
		from,
		client,
		[]string{"(address,bytes,bool)[]"},
		blockNumber,
		overrides,
	)
	if err != nil {
		return Result{Success: false, Error: err, TxOrCall: txOrCall}
	}

	resultArgs, err := abi.Decode([]string{"(bool,bytes)[]"}, common.Hex2Bytes(rawResponse[2:]))
	if err != nil {
		return Result{Success: false, Error: err, TxOrCall: txOrCall}
	}
	resultArgs = resultArgs[0].([]any)

	var result []any
	for i, call := range calls {
		resultArgs[i].([]any)[1], err = abi.Decode(call.ReturnTypes, resultArgs[i].([]any)[1].([]byte))
		if err != nil {
			return Result{Success: false, Error: err, TxOrCall: txOrCall}
		}

		result = append(result, resultArgs[i])
	}

	return Result{Success: true, Result: result, TxOrCall: txOrCall}
}

func deploylessGetCodeLengths(
	addresses []*common.Address, client *ethclient.Client, blockNumber *big.Int,
) Result {

	rawResponse, txOrCall, err := makeDeploylessCall(
		toAnyArray(addresses), false, CODE_LENGTH, nil, client, []string{"address[]"}, blockNumber, nil,
	)
	if err != nil {
		return Result{Success: false, Error: err, TxOrCall: txOrCall}
	}

	resultArgs, err := abi.Decode([]string{"uint256[]"}, common.Hex2Bytes(rawResponse[2:]))
	if err != nil {
		return Result{Success: false, Error: err, TxOrCall: txOrCall}
	}
	resultArgs = resultArgs[0].([]any)

	return Result{Success: true, Result: resultArgs, TxOrCall: txOrCall}
}

func deploylessGetBalances(
	addresses []*common.Address, client *ethclient.Client, blockNumber *big.Int,
) Result {

	rawResponse, txOrCall, err := makeDeploylessCall(
		toAnyArray(addresses), false, BALANCES, nil, client, []string{"address[]"}, blockNumber, nil,
	)
	if err != nil {
		return Result{Success: false, Error: err, TxOrCall: txOrCall}
	}

	resultArgs, err := abi.Decode([]string{"uint256[]"}, common.Hex2Bytes(rawResponse[2:]))
	if err != nil {
		return Result{Success: false, Error: err, TxOrCall: txOrCall}
	}
	resultArgs = resultArgs[0].([]any)

	return Result{Success: true, Result: resultArgs, TxOrCall: txOrCall}
}

func deploylessGetAddressesData(
	addresses []*common.Address, client *ethclient.Client, blockNumber *big.Int,
) Result {

	rawResponse, txOrCall, err := makeDeploylessCall(
		toAnyArray(addresses), false, ADDRESSES_DATA, nil, client, []string{"address[]"}, blockNumber, nil,
	)
	if err != nil {
		return Result{Success: false, Error: err, TxOrCall: txOrCall}
	}

	resultArgs, err := abi.Decode([]string{"uint256[]", "uint256[]"}, common.Hex2Bytes(rawResponse[2:]))
	if err != nil {
		return Result{Success: false, Error: err, TxOrCall: txOrCall}
	}

	var result [][]any
	for i := range addresses {
		result = append(result, []any{resultArgs[0].([]any)[i], resultArgs[1].([]any)[i]})
	}

	return Result{Success: true, Result: result, TxOrCall: txOrCall}
}

func deploylessGetChainData(client *ethclient.Client, blockNumber *big.Int) Result {

	rawResponse, txOrCall, err := makeDeploylessCall(
		nil, false, CHAIN_DATA, nil, client, nil, blockNumber, nil,
	)
	if err != nil {
		return Result{Success: false, Error: err, TxOrCall: txOrCall}
	}

	resultArgs, err := abi.Decode(
		[]string{
			"uint256",
			"uint256",
			"bytes32",
			"uint256",
			"address",
			"uint256",
			"uint256",
			"uint256",
			"uint256",
		},
		common.Hex2Bytes(rawResponse[2:]),
	)
	if err != nil {
		return Result{Success: false, Error: err, TxOrCall: txOrCall}
	}

	return Result{Success: true, Result: resultArgs, TxOrCall: txOrCall}
}

func makeDeploylessCall(
	params []any, requireSuccess bool, callType CallType,
	from *common.Address, client *ethclient.Client, typeStrs []string, blockNumber *big.Int, overrides StateOverride,
) (string, TxOrCall, error) {
	var encoded []byte
	var err error
	if callType == TRY_STATIC_CALL {
		encoded, err = abi.Encode(typeStrs, params, requireSuccess)
	} else if typeStrs != nil && params != nil {
		encoded, err = abi.Encode(typeStrs, params)
	}
	if err != nil {
		return "", TxOrCall{}, err
	}

	encodedParams, err := abi.EncodePacked([]string{"uint8", "bytes"}, big.NewInt(int64(callType)), encoded)
	if err != nil {
		return "", TxOrCall{}, err
	}

	encodedParamsToDeploy, err := abi.Encode([]string{"bytes"}, encodedParams)
	if err != nil {
		return "", TxOrCall{}, err
	}

	data := DEPLOYLESS_MULTICALL_BYTECODE + common.Bytes2Hex(encodedParamsToDeploy)

	var blockIdentifier string
	if blockNumber != nil {
		blockIdentifier = hexutil.EncodeBig(blockNumber)
	} else {
		blockIdentifier = "latest"
	}

	var call CallArgs
	if from == nil {
		call = CallArgs{
			To:   nil,
			Data: hexutil.Bytes(data),
		}
	} else {
		call = CallArgs{
			From: *from,
			To:   nil,
			Data: hexutil.Bytes(data),
		}
	}

	var rawResponse string
	err = client.Client().CallContext(context.Background(), &rawResponse, "eth_call", call, blockIdentifier, overrides)
	if err != nil {
		return rawResponse, TxOrCall{}, fmt.Errorf("error making deployless call: %w, with data: %s", err, data)
	}

	if blockNumber == nil {
		blockNumberUint64, err := client.BlockNumber(context.Background())
		if err != nil {
			return rawResponse, TxOrCall{}, fmt.Errorf("error getting block number: %w", err)
		}
		blockNumber = big.NewInt(int64(blockNumberUint64))
	}

	return rawResponse, TxOrCall{To: nil, Data: common.Hex2Bytes(data), BlockNumber: blockNumber}, nil
}

func toAnyArray(addresses []*common.Address) []any {

	var anyUserOps []any
	for _, address := range addresses {
		anyUserOps = append(anyUserOps, address)
	}

	return anyUserOps
}
