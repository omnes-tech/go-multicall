package multicall

import (
	"context"
	"fmt"
	"log"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type MultiCall struct {
	ContractAddress *common.Address
	Signer          *SignerInterface
}

func NewMultiCall(client *ethclient.Client, signer *SignerInterface) (*MultiCall, error) {
	var multicallAddress *common.Address

	bytecode, err := client.CodeAt(context.Background(), OMNES_MULTICALL_ADDRESS, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting bytecode: %v", err)
	}

	if len(bytecode) == 0 {
		log.Printf("no deployed contract found. Using deployless method\n\n")

		multicallAddress = nil
	} else {
		multicallAddress = &OMNES_MULTICALL_ADDRESS
	}

	return &MultiCall{
		ContractAddress: multicallAddress,
		Signer:          signer,
	}, nil

}

func (m *MultiCall) AggregateCalls(
	calls []Call, client *ethclient.Client, blockNumber *big.Int, isCall bool,
) Result {
	if m.Signer == nil && !isCall {
		return Result{Success: false, Error: fmt.Errorf("no signer configured")}
	}
	if m.ContractAddress == nil {
		return Result{Success: false, Error: fmt.Errorf("no multicall contract on this chain")}
	}

	if isCall {
		return txAsRead(
			calls,
			false,
			client,
			m.ContractAddress,
			"aggregateCalls((address,bytes,uint256)[])",
			[]string{"bytes[]"},
			blockNumber,
		)
	} else {
		return transact(
			calls,
			false,
			client,
			*m.Signer,
			m.ContractAddress,
			"aggregateCalls((address,bytes,uint256)[])",
			[]string{"bytes[]"},
			true,
			false,
		)
	}

}

func (m *MultiCall) TryAggregateCalls(
	calls []Call, requireSuccess bool, client *ethclient.Client, blockNumber *big.Int, isCall bool,
) Result {
	if m.Signer == nil && !isCall {
		return Result{Success: false, Error: fmt.Errorf("no signer configured")}
	}
	if m.ContractAddress == nil {
		return Result{Success: false, Error: fmt.Errorf("no multicall contract on this chain")}
	}

	if isCall {
		return txAsRead(
			calls,
			requireSuccess,
			client,
			m.ContractAddress,
			"tryAggregateCalls((address,bytes,uint256)[],bool)",
			[]string{"(bool,bytes)[]"},
			blockNumber,
		)
	} else {
		return transact(
			calls,
			requireSuccess,
			client,
			*m.Signer,
			m.ContractAddress,
			"tryAggregateCalls((address,bytes,uint256)[],bool)",
			[]string{"(bool,bytes)[]"},
			true,
			false,
		)
	}

}

func (m *MultiCall) TryAggregateCalls3(
	calls []CallWithFailure, client *ethclient.Client, blockNumber *big.Int, isCall bool,
) Result {
	if m.Signer == nil && !isCall {
		return Result{Success: false, Error: fmt.Errorf("no signer configured")}
	}
	if m.ContractAddress == nil {
		return Result{Success: false, Error: fmt.Errorf("no multicall contract on this chain")}
	}

	if isCall {
		return txAsReadWithFailure(
			calls,
			false,
			client,
			m.ContractAddress,
			"tryAggregateCalls((address,bytes,uint256,bool)[])",
			[]string{"(bool,bytes)[]"},
			blockNumber,
		)
	} else {
		return transactWithFailure(
			calls,
			false,
			client,
			*m.Signer,
			m.ContractAddress,
			"tryAggregateCalls((address,bytes,uint256,bool)[])",
			[]string{"(bool,bytes)[]"},
			true,
			false,
		)
	}

}

func (m *MultiCall) SimulateCall(
	calls []Call, client *ethclient.Client, blockNumber *big.Int,
) Result {
	if m.ContractAddress == nil {
		return deploylessSimulation(calls, client, blockNumber)
	}

	return call(
		calls,
		false,
		client,
		m.ContractAddress,
		"simulateCalls((address,bytes)[])",
		nil,
		m.ContractAddress,
		blockNumber,
		true,
	)

}

func (m *MultiCall) AggregateStatic(
	calls []Call, client *ethclient.Client, blockNumber *big.Int,
) Result {
	if m.ContractAddress == nil {
		return deploylessAggregateStatic(calls, client, blockNumber)
	}

	return call(
		calls,
		false,
		client,
		m.ContractAddress,
		"aggregateStatic((address,bytes)[])",
		[]string{"bytes[]"},
		m.ContractAddress,
		blockNumber,
		false,
	)

}

func (m *MultiCall) TryAggregateStatic(
	calls []Call, requireSuccess bool, client *ethclient.Client, blockNumber *big.Int,
) Result {
	if m.ContractAddress == nil {
		return deploylessTryAggregateStatic(calls, requireSuccess, client, blockNumber)
	}

	return call(
		calls,
		requireSuccess,
		client,
		m.ContractAddress,
		"tryAggregateStatic((address,bytes)[],bool)",
		[]string{"(bool,bytes)[]"},
		m.ContractAddress,
		blockNumber,
		false,
	)

}

func (m *MultiCall) TryAggregateStatic3(
	calls []CallWithFailure, client *ethclient.Client, blockNumber *big.Int,
) Result {
	if m.ContractAddress == nil {
		return deploylessTryAggregateStatic3(calls, client, blockNumber)
	}

	return callWithFailure(
		calls,
		client,
		m.ContractAddress,
		"tryAggregateStatic((address,bytes,bool)[])",
		[]string{"(bool,bytes)[]"},
		m.ContractAddress,
		blockNumber,
	)

}

func (m *MultiCall) CodeLengths(
	addresses []*common.Address, client *ethclient.Client, blockNumber *big.Int,
) Result {
	if m.ContractAddress == nil {
		return deploylessGetCodeLengths(addresses, client, blockNumber)
	}

	return getData(
		addresses,
		client,
		m.ContractAddress,
		"getCodeLengths(address[])",
		[]string{"uint256[]"},
		blockNumber,
	)

}

func (m *MultiCall) Balances(
	addresses []*common.Address, client *ethclient.Client, blockNumber *big.Int,
) Result {
	if m.ContractAddress == nil {
		return deploylessGetBalances(addresses, client, blockNumber)
	}

	return getData(
		addresses,
		client,
		m.ContractAddress,
		"getBalances(address[])",
		[]string{"uint256[]"},
		blockNumber,
	)
}

func (m *MultiCall) AddressesData(
	addresses []*common.Address, client *ethclient.Client, blockNumber *big.Int,
) Result {
	if m.ContractAddress == nil {
		return deploylessGetAddressesData(addresses, client, blockNumber)
	}

	return getData(
		addresses,
		client,
		m.ContractAddress,
		"getAddressesData(address[])",
		[]string{"uint256[]", "uint256[]"},
		blockNumber,
	)
}

func (m *MultiCall) ChainData(client *ethclient.Client, blockNumber *big.Int) Result {
	if m.ContractAddress == nil {
		return deploylessGetChainData(client, blockNumber)
	}

	return getData(
		nil,
		client,
		m.ContractAddress,
		"getChainData()",
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
		blockNumber,
	)
}

// IsDeployed checks if the multicall contract is deployed on the chain.
func (m *MultiCall) IsDeployed() bool {
	return m.ContractAddress != nil
}

// IsDeployless checks if the multicall contract is deployless.
func (m *MultiCall) IsDeployless() bool {
	return m.ContractAddress == nil
}
