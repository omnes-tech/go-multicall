package multicall

import (
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/omnes-tech/abi"
)

type TxOrCall struct {
	From        common.Address
	To          *common.Address
	Gas         uint64
	GasPrice    *big.Int
	GasFeeCap   *big.Int
	GasTipCap   *big.Int
	Value       *big.Int
	Data        []byte
	Nonce       uint64
	BlockNumber *big.Int

	AccessList types.AccessList
	Overrides  StateOverride
}

func (t *TxOrCall) String() string {
	return fmt.Sprintf(`TxOrCall{
	From: %s, 
	To: %s, 
	Gas: %d, 
	GasPrice: %s, 
	Value: %s, 
	Data: %s,
	Nonce: %d,
	BlockNumber: %s,
	AccessList: %v,
	Overrides: %v,
}
`,
		t.From.Hex(),
		t.To.Hex(),
		t.Gas,
		t.GasPrice.String(),
		t.Value.String(),
		common.Bytes2Hex(t.Data),
		t.Nonce,
		t.BlockNumber.String(),
		t.AccessList,
		t.Overrides,
	)
}

func FromTxToTxOrCall(tx *types.Transaction, from common.Address, blockNumber *big.Int, overrides StateOverride) TxOrCall {
	return TxOrCall{
		From:        from,
		To:          tx.To(),
		Gas:         tx.Gas(),
		GasPrice:    tx.GasPrice(),
		Value:       tx.Value(),
		Data:        tx.Data(),
		Nonce:       tx.Nonce(),
		AccessList:  tx.AccessList(),
		BlockNumber: blockNumber,
		Overrides:   overrides,
	}
}

func FromCallToTxOrCall(call *ethereum.CallMsg, blockNumber *big.Int, overrides StateOverride) TxOrCall {
	return TxOrCall{
		From:        call.From,
		To:          call.To,
		Gas:         call.Gas,
		GasPrice:    call.GasPrice,
		Value:       call.Value,
		Data:        call.Data,
		BlockNumber: blockNumber,
		Overrides:   overrides,
	}
}

type Result struct {
	Success  bool
	Result   any
	Error    error
	TxOrCall TxOrCall
}

func (r *Result) Description(full bool) string {
	if full {
		return fmt.Sprintf("%+v", r)
	} else {
		return fmt.Sprintf("Success: %+v, Result: %+v", r.Success, r.Result)
	}
}

func (r *Result) Print(full bool) {
	fmt.Println(r.Description(full))
}

type commonCall struct {
	Target        common.Address
	FuncSignature string
	Args          []interface{}
	ReturnTypes   []string
	CallData      []byte
}

type Call struct {
	commonCall
	Value *big.Int
}

type CallWithFailure struct {
	Call
	RequireSuccess bool
}

type StaticCall commonCall

type StaticCallWithFailure struct {
	StaticCall
	RequireSuccess bool
}

type CallsInterface interface {
	GetTarget(i int) *common.Address
	GetFuncSignature(i int) string
	GetArgs(i int) []interface{}
	GetCallData(i int) []byte
	GetReturnTypes(i int) []string
	GetValue(i int) *big.Int
	Len() int
	ToArray(withValue bool, isMultiCall3Type bool) ([]any, *big.Int, error)
}

type CallsWithFailureInterface interface {
	CallsInterface
	GetRequireSuccess(i int) bool
}

type Calls []Call
type CallsWithFailure []CallWithFailure

func NewCall(
	target common.Address, funcSignature string,
	args []interface{}, callData []byte, returnTypes []string, value *big.Int,
) Call {
	return Call{
		commonCall: commonCall{
			Target:        target,
			FuncSignature: funcSignature,
			Args:          args,
			ReturnTypes:   returnTypes,
			CallData:      callData,
		},
		Value: value,
	}
}

func NewCalls(
	targets []common.Address, funcSignatures []string,
	argss [][]interface{}, callDatas [][]byte, returnTypess [][]string, values []*big.Int,
) Calls {
	calls := make(Calls, len(targets))
	for i, target := range targets {
		args, callData, returnTypes, value := setParameters(i, argss, callDatas, returnTypess, values)

		if funcSignatures == nil {
			calls[i] = NewCall(target, "", args, callData, returnTypes, value)
		} else {
			calls[i] = NewCall(target, funcSignatures[i], args, callData, returnTypes, value)
		}
	}

	return calls
}

func NewCallWithFailure(
	target common.Address, funcSignature string, args []interface{},
	callData []byte, returnTypes []string, value *big.Int, requireSuccess bool,
) CallWithFailure {
	return CallWithFailure{
		Call: Call{
			commonCall: commonCall{
				Target:        target,
				FuncSignature: funcSignature,
				Args:          args,
				ReturnTypes:   returnTypes,
			},
			Value: value,
		},
		RequireSuccess: requireSuccess,
	}
}

func NewCallsWithFailure(
	targets []common.Address, funcSignatures []string, argss [][]interface{},
	callDatas [][]byte, returnTypess [][]string, values []*big.Int, requireSuccesss []bool,
) CallsWithFailure {
	calls := make(CallsWithFailure, len(targets))
	for i, target := range targets {
		args, callData, returnTypes, value := setParameters(i, argss, callDatas, returnTypess, values)

		if funcSignatures == nil {
			calls[i] = NewCallWithFailure(target, "", args, callData, returnTypes, value, requireSuccesss[i])
		} else {
			calls[i] = NewCallWithFailure(target, funcSignatures[i], args, callData, returnTypes, value, requireSuccesss[i])
		}
	}

	return calls
}

func ParseCallToCalls(calls []Call) Calls {
	result := make(Calls, len(calls))
	for i, c := range calls {
		result[i] = Call{
			commonCall: commonCall{
				Target:        c.Target,
				FuncSignature: c.FuncSignature,
				Args:          c.Args,
				ReturnTypes:   c.ReturnTypes,
			},
			Value: c.Value,
		}
	}
	return result
}

func ParseCallWithFailureToCallsWithFailure(calls []CallWithFailure) CallsWithFailure {
	result := make(CallsWithFailure, len(calls))
	for i, c := range calls {
		result[i] = CallWithFailure{
			Call: Call{
				commonCall: commonCall{
					Target:        c.Target,
					FuncSignature: c.FuncSignature,
					Args:          c.Args,
					ReturnTypes:   c.ReturnTypes,
				},
				Value: c.Value,
			},
			RequireSuccess: c.RequireSuccess,
		}
	}
	return result
}

func setParameters(
	index int, argss [][]interface{}, callDatas [][]byte, returnTypess [][]string, values []*big.Int,
) ([]any, []byte, []string, *big.Int) {
	var args []any
	if argss == nil {
		args = nil
	} else {
		args = argss[index]
	}

	var callData []byte
	if callDatas == nil {
		callData = nil
	} else {
		callData = callDatas[index]
	}

	var returnTypes []string
	if returnTypess == nil {
		returnTypes = nil
	} else {
		returnTypes = returnTypess[index]
	}

	var value *big.Int
	if values == nil {
		value = nil
	} else {
		value = values[index]
	}

	return args, callData, returnTypes, value
}

func (c Calls) GetTarget(i int) *common.Address {
	return &c[i].Target
}

func (c Calls) GetFuncSignature(i int) string {
	return c[i].FuncSignature
}

func (c Calls) GetArgs(i int) []interface{} {
	return c[i].Args
}

func (c Calls) GetCallData(i int) []byte {
	return c[i].CallData
}

func (c Calls) GetReturnTypes(i int) []string {
	return c[i].ReturnTypes
}

func (c Calls) GetValue(i int) *big.Int {
	return c[i].Value
}

func (c Calls) Len() int {
	return len(c)
}

func (c Calls) ToArray(withValue bool, isMultiCall3Type bool) ([]any, *big.Int, error) {
	var result []any
	summed := big.NewInt(0)
	for i := 0; i < c.Len(); i++ {
		var args []any
		args = append(args, c.GetTarget(i))

		callData := c.GetCallData(i)
		if callData == nil {
			var err error
			if c.GetArgs(i) != nil || len(c.GetArgs(i)) > 0 {
				callData, err = abi.EncodeWithSignature(c.GetFuncSignature(i), c.GetArgs(i)...)
			} else {
				callData, err = abi.EncodeWithSignature(c.GetFuncSignature(i))
			}
			if err != nil {
				return nil, nil, err
			}
		}
		args = append(args, callData)

		if withValue {
			value := big.NewInt(0)
			if c.GetValue(i) != nil {
				value.Add(value, c.GetValue(i))
			}
			summed.Add(summed, value)

			args = append(args, value)
		}

		result = append(result, args)
	}

	return result, summed, nil
}

func (c CallsWithFailure) GetTarget(i int) *common.Address {
	return &c[i].Target
}

func (c CallsWithFailure) GetFuncSignature(i int) string {
	return c[i].FuncSignature
}

func (c CallsWithFailure) GetArgs(i int) []interface{} {
	return c[i].Args
}

func (c CallsWithFailure) GetCallData(i int) []byte {
	return c[i].CallData
}

func (c CallsWithFailure) GetReturnTypes(i int) []string {
	return c[i].ReturnTypes
}

func (c CallsWithFailure) GetValue(i int) *big.Int {
	return c[i].Value
}

func (c CallsWithFailure) GetRequireSuccess(i int) bool {
	return c[i].RequireSuccess
}

func (c CallsWithFailure) Len() int {
	return len(c)
}

func (c CallsWithFailure) ToArray(withValue bool, isMultiCall3Type bool) ([]any, *big.Int, error) {
	// Omnes: (address,bytes,uint256,bool)
	// MultiCall3: (address,bool,bytes) or (address,bool,uint256,bytes)

	var result []any
	summed := big.NewInt(0)
	for i := 0; i < c.Len(); i++ {
		var args []any

		callData := c.GetCallData(i)
		if callData == nil {
			var err error
			if c.GetArgs(i) != nil || len(c.GetArgs(i)) > 0 {
				callData, err = abi.EncodeWithSignature(c.GetFuncSignature(i), c.GetArgs(i)...)
			} else {
				callData, err = abi.EncodeWithSignature(c.GetFuncSignature(i))
			}
			if err != nil {
				return nil, nil, err
			}
		}

		if isMultiCall3Type {
			args = append(args, c.GetTarget(i))
			args = append(args, c.GetRequireSuccess(i))
			if withValue {
				value := big.NewInt(0)
				if c.GetValue(i) != nil {
					value.Add(value, c.GetValue(i))
				}
				summed.Add(summed, value)

				args = append(args, value)
			}
			args = append(args, callData)
		} else {
			args = append(args, c.GetTarget(i))
			args = append(args, callData)
			if withValue {
				value := big.NewInt(0)
				if c.GetValue(i) != nil {
					value.Add(value, c.GetValue(i))
				}
				summed.Add(summed, value)

				args = append(args, value)
			}
			args = append(args, c.GetRequireSuccess(i))
		}

		result = append(result, args)
	}

	return result, summed, nil
}

// StateOverride matches the shape geth expects for eth_call overrides
// StateDiff: patch specific storage slots, leave rest untouched
// State: replace entire storage, unset slots become zero
// Balance: override ETH balance
// Nonce: override account nonce
// Code: replace contract bytecode entirely
type OverrideAccount struct {
	StateDiff map[common.Hash]common.Hash `json:"stateDiff,omitempty"`
	State     map[common.Hash]common.Hash `json:"state,omitempty"`
	Balance   *hexutil.Big                `json:"balance,omitempty"`
	Nonce     *hexutil.Uint64             `json:"nonce,omitempty"`
	Code      hexutil.Bytes               `json:"code,omitempty"`
}

type StateOverride map[common.Address]OverrideAccount

type Overrides struct {
	From           *common.Address
	StateOverrides StateOverride
	BlockNumber    *big.Int
}

// CallMsg-equivalent as a raw map that handles JSON-marshaled RPC data
type CallArgs struct {
	From  common.Address  `json:"from,omitempty"`
	To    *common.Address `json:"to,omitempty"`
	Data  hexutil.Bytes   `json:"data,omitempty"`
	Value *big.Int        `json:"value,omitempty"`
}

func (c *CallArgs) ToEthereumCallMsg() *ethereum.CallMsg {
	return &ethereum.CallMsg{
		From:  c.From,
		To:    c.To,
		Data:  c.Data,
		Value: c.Value,
	}
}
