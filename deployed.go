package multicall

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/omnes-tech/abi"
)

func transactWithFailure(
	calls CallsWithFailure, requireSuccess bool, client *ethclient.Client,
	signer SignerInterface, to *common.Address, funcSignature string, txReturnTypes []string,
	withValue bool, isMultiCall3Type bool,
) Result {
	return write(
		calls,
		requireSuccess,
		client,
		signer,
		to,
		funcSignature,
		txReturnTypes,
		withValue,
		isMultiCall3Type,
	)
}

func transact(
	calls Calls, requireSuccess bool, client *ethclient.Client,
	signer SignerInterface, to *common.Address, funcSignature string, txReturnTypes []string,
	withValue bool, isMultiCall3Type bool,
) Result {
	return write(
		calls,
		requireSuccess,
		client,
		signer,
		to,
		funcSignature,
		txReturnTypes,
		withValue,
		isMultiCall3Type,
	)
}

func write(
	calls CallsInterface, requireSuccess bool, client *ethclient.Client, signer SignerInterface,
	to *common.Address, funcSignature string, txReturnTypes []string, withValue bool, isMultiCall3Type bool,
) Result {
	arrayfiedCalls, msgValue, err := calls.ToArray(withValue, isMultiCall3Type)
	if err != nil {
		return Result{Success: false, Error: err}
	}

	var callData []byte
	if funcSignature == "tryAggregateCalls((address,bytes,uint256)[],bool)" {
		callData, err = abi.EncodeWithSignature(funcSignature, arrayfiedCalls, requireSuccess)
	} else {
		callData, err = abi.EncodeWithSignature(funcSignature, arrayfiedCalls)
	}
	if err != nil {
		return Result{Success: false, Error: err}
	}

	tx, err := createTransaction(client, signer.GetAddress(), to, msgValue, callData)
	if err != nil {
		return Result{Success: false, Error: err, TxOrCall: FromTxToTxOrCall(tx, *signer.GetAddress(), nil)}
	}

	chainId, err := client.ChainID(context.Background())
	if err != nil {
		return Result{Success: false, Error: err, TxOrCall: FromTxToTxOrCall(tx, *signer.GetAddress(), nil)}
	}

	signedTx, err := signer.SignTx(tx, chainId)
	if err != nil {
		return Result{Success: false, Error: err, TxOrCall: FromTxToTxOrCall(tx, *signer.GetAddress(), nil)}
	}

	encodedCallResult, err := client.CallContract(context.Background(), ethereum.CallMsg{
		From: *signer.GetAddress(),
		To:   to,
		Data: callData,
	}, nil)
	if err != nil {
		blockNumber, err := client.BlockNumber(context.Background())
		if err != nil {
			return Result{Success: false, Error: err, TxOrCall: FromTxToTxOrCall(tx, *signer.GetAddress(), nil)}
		}

		return Result{
			Success:  false,
			Error:    fmt.Errorf("error calling contract: %w, with data: %s", err, common.Bytes2Hex(callData)),
			TxOrCall: FromTxToTxOrCall(tx, *signer.GetAddress(), big.NewInt(int64(blockNumber))),
		}
	}

	receipt, err := sendSignedTransaction(client, signedTx)
	if err != nil {
		return Result{
			Success:  false,
			Error:    fmt.Errorf("error sending signed transaction: %w", err),
			TxOrCall: FromTxToTxOrCall(tx, *signer.GetAddress(), receipt.BlockNumber),
		}
	}

	decodedCallResult, err := abi.Decode(txReturnTypes, encodedCallResult)
	if err != nil {
		return Result{
			Success:  false,
			Error:    fmt.Errorf("error decoding call result: %w", err),
			TxOrCall: FromTxToTxOrCall(tx, *signer.GetAddress(), receipt.BlockNumber),
		}
	}

	return parseResults(decodedCallResult, receipt.Status == 1, receipt, FromTxToTxOrCall(signedTx, *signer.GetAddress(), receipt.BlockNumber))
}

func txAsReadWithFailure(
	calls CallsWithFailure, requireSuccess bool, client *ethclient.Client, to *common.Address,
	funcSignature string, txReturnTypes []string, blockNumber *big.Int,
) Result {
	return asRead(
		calls,
		requireSuccess,
		client,
		to,
		funcSignature,
		txReturnTypes,
		blockNumber,
	)
}

func txAsRead(
	calls Calls, requireSuccess bool, client *ethclient.Client, to *common.Address,
	funcSignature string, txReturnTypes []string, blockNumber *big.Int,
) Result {
	return asRead(
		calls,
		requireSuccess,
		client,
		to,
		funcSignature,
		txReturnTypes,
		blockNumber,
	)
}

func asRead(
	calls CallsInterface, requireSuccess bool, client *ethclient.Client, to *common.Address,
	funcSignature string, txReturnTypes []string, blockNumber *big.Int,
) Result {
	arrayfiedCalls, _, err := calls.ToArray(true, false)
	if err != nil {
		return Result{Success: false, Error: err}
	}

	var callData []byte
	if funcSignature == "tryAggregateCalls((address,bytes,uint256)[],bool)" {
		callData, err = abi.EncodeWithSignature(funcSignature, arrayfiedCalls, requireSuccess)
	} else {
		callData, err = abi.EncodeWithSignature(funcSignature, arrayfiedCalls)
	}
	if err != nil {
		return Result{Success: false, Error: err}
	}

	decodedCallResult, decodedAggregatedCallsResultVar, call, err := makeCall(
		calls,
		client,
		to,
		callData,
		txReturnTypes,
		false,
		nil,
		blockNumber,
	)
	if err != nil {
		return Result{Success: false, Error: err, TxOrCall: call}
	}

	return parseResults(decodedAggregatedCallsResultVar, true, decodedCallResult, call)
}

func call(
	calls Calls, requireSuccess bool, client *ethclient.Client, to *common.Address, funcSignature string,
	txReturnTypes []string, multicallAddress *common.Address,
	blockNumber *big.Int, isSimulation bool,
) Result {
	return read(
		calls,
		requireSuccess,
		client,
		to,
		funcSignature,
		txReturnTypes,
		multicallAddress,
		blockNumber,
		isSimulation,
	)
}

func callWithFailure(
	calls CallsWithFailure, client *ethclient.Client, to *common.Address, funcSignature string,
	txReturnTypes []string, multicallAddress *common.Address, blockNumber *big.Int,
) Result {
	return read(
		calls,
		false,
		client,
		to,
		funcSignature,
		txReturnTypes,
		multicallAddress,
		blockNumber,
		false,
	)
}

func read(
	calls CallsInterface, requireSuccess bool, client *ethclient.Client, to *common.Address, funcSignature string,
	txReturnTypes []string, multicallAddress *common.Address, blockNumber *big.Int,
	isSimulation bool,
) Result {
	arrayfiedCalls, _, err := calls.ToArray(false, false)
	if err != nil {
		return Result{Success: false, Error: err}
	}

	if funcSignature == "tryAggregateStatic((address,bytes,bool)[])" {
		isSimulation = false
	}

	var callData []byte
	if funcSignature == "tryAggregateStatic((address,bytes)[],bool)" {
		isSimulation = false
		callData, err = abi.EncodeWithSignature(funcSignature, arrayfiedCalls, requireSuccess)
	} else {
		callData, err = abi.EncodeWithSignature(funcSignature, arrayfiedCalls)
	}
	if err != nil {
		return Result{Success: false, Error: err}
	}

	decodedCallResult, decodedAggregatedCallsResultVar, call, err := makeCall(
		calls,
		client,
		to,
		callData,
		txReturnTypes,
		isSimulation,
		multicallAddress,
		blockNumber,
	)
	if err != nil {
		return Result{Success: false, Error: err, TxOrCall: call}
	}

	return parseResults(decodedAggregatedCallsResultVar, true, decodedCallResult, call)
}

func getData(
	addresses []*common.Address, client *ethclient.Client, to *common.Address,
	funcSignature string, returnTypes []string, blockNumber *big.Int,
) Result {

	var callData []byte
	var err error
	if addresses != nil {
		callData, err = abi.EncodeWithSignature(funcSignature, toAnyArray(addresses))
	} else {
		callData, err = abi.EncodeWithSignature(funcSignature)
	}
	if err != nil {
		return Result{Success: false, Error: err}
	}

	encodedCallResult, call, err := readContract(client, &ZERO_ADDRESS, to, callData, blockNumber)
	if err != nil {
		return Result{Success: false, Error: err, TxOrCall: FromCallToTxOrCall(call, blockNumber)}
	}

	decodedCallResult, err := abi.Decode(returnTypes, encodedCallResult)
	if err != nil {
		return Result{Success: false, Error: err, TxOrCall: FromCallToTxOrCall(call, blockNumber)}
	}

	if blockNumber == nil {
		blockNumberUint64, err := client.BlockNumber(context.Background())
		if err != nil {
			return Result{Success: false, Error: err, TxOrCall: FromCallToTxOrCall(call, blockNumber)}
		}
		blockNumber = big.NewInt(int64(blockNumberUint64))
	}

	return Result{Success: true, Result: decodedCallResult, TxOrCall: FromCallToTxOrCall(call, blockNumber)}
}

func makeCall(
	calls CallsInterface, client *ethclient.Client, to *common.Address, callData []byte, txReturnTypes []string,
	isSimulation bool, multicallAddress *common.Address, blockNumber *big.Int,
) ([]any, []any, TxOrCall, error) {
	if !true {
		log.Println(multicallAddress)
	}

	var decodedCallResult []any
	encodedCallResult, call, err := readContract(client, &ZERO_ADDRESS, to, callData, blockNumber)
	if err != nil && !isSimulation {
		return nil, nil, TxOrCall{}, err
	} else if isSimulation {
		if strings.Contains(err.Error(), "execution reverted") {
			encodedRevert, ok := parseRevertData(err)
			if ok {
				decodedCallResult, err := abi.DecodeWithSignature(
					"MultiCall__Simulation((bool,bytes,uint256)[])",
					encodedRevert,
				)
				if err != nil {
					return nil, nil, TxOrCall{}, err
				}
				decodedCallResult = decodedCallResult[0].([]any)

				for i, result := range decodedCallResult {
					decodedCallResult[i].([]any)[1] = common.Bytes2Hex(result.([]any)[1].([]byte))
				}
			}
		}
	}
	if len(encodedCallResult) == 0 {
		multicallAddress = nil
	}

	if !isSimulation {
		decodedCallResult, err = abi.Decode(txReturnTypes, encodedCallResult)
		if err != nil {
			return nil, nil, TxOrCall{}, err
		}
	}

	for len(decodedCallResult) != calls.Len() {
		decodedCallResult = decodedCallResult[0].([]any)
	}

	decodedAggregatedCallsResultVar, err := decodeAggregateCallsResult(decodedCallResult, calls)
	if err != nil {
		return nil, nil, TxOrCall{}, err
	}

	if blockNumber == nil {
		blockNumberUint64, err := client.BlockNumber(context.Background())
		if err != nil {
			return nil, nil, TxOrCall{}, err
		}
		blockNumber = big.NewInt(int64(blockNumberUint64))
	}

	return decodedCallResult, decodedAggregatedCallsResultVar, FromCallToTxOrCall(call, blockNumber), nil
}

func parseResults(
	decodedCallResult []any, status bool, callOrTxResult any, callOrTx TxOrCall,
) Result {
	var result Result
	if len(decodedCallResult) > 0 {
		result = Result{
			Success:  status,
			Result:   decodedCallResult,
			Error:    nil,
			TxOrCall: callOrTx,
		}
	} else {
		result = Result{
			Success:  status,
			Result:   callOrTxResult,
			Error:    nil,
			TxOrCall: callOrTx,
		}
	}

	return result
}

func decodeAggregateCallsResult(result []any, calls CallsInterface) ([]any, error) {
	var decodedResult []any
	for i, res := range result {
		returnTypes := calls.GetReturnTypes(i)
		if returnTypes != nil || len(returnTypes) > 0 {
			r, ok := res.([]byte)
			if ok {
				decodedR, err := abi.Decode(returnTypes, r)
				if err != nil {
					return nil, err
				}

				decodedResult = append(decodedResult, decodedR)
			} else {
				decodedResult = append(decodedResult, res.([]any))
			}
		} else {
			decodedResult = append(decodedResult, res.([]any))
		}
	}

	return decodedResult, nil
}
