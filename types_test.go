package multicall

import (
	"bytes"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func TestOverrideAccount_Add_Balance(t *testing.T) {
	t.Run("sums when both set", func(t *testing.T) {
		o := OverrideAccount{Balance: balanceHex(100)}
		o.Add(OverrideAccount{Balance: balanceHex(50)})

		if got := o.Balance.ToInt().Int64(); got != 150 {
			t.Fatalf("Balance = %d, want 150", got)
		}
	})

	t.Run("keeps receiver when other unset", func(t *testing.T) {
		o := OverrideAccount{Balance: balanceHex(100)}
		o.Add(OverrideAccount{})

		if o.Balance == nil || o.Balance.ToInt().Int64() != 100 {
			t.Fatalf("Balance = %v, want 100", o.Balance)
		}
	})

	t.Run("keeps other when receiver unset", func(t *testing.T) {
		o := OverrideAccount{}
		o.Add(OverrideAccount{Balance: balanceHex(75)})

		if o.Balance == nil || o.Balance.ToInt().Int64() != 75 {
			t.Fatalf("Balance = %v, want 75", o.Balance)
		}
	})
}

func TestOverrideAccount_Add_Nonce(t *testing.T) {
	t.Run("other wins when both set", func(t *testing.T) {
		n1 := hexutil.Uint64(1)
		n2 := hexutil.Uint64(2)
		o := OverrideAccount{Nonce: &n1}
		o.Add(OverrideAccount{Nonce: &n2})

		if o.Nonce == nil || uint64(*o.Nonce) != 2 {
			t.Fatalf("Nonce = %v, want 2", o.Nonce)
		}
	})

	t.Run("keeps receiver when other unset", func(t *testing.T) {
		n1 := hexutil.Uint64(1)
		o := OverrideAccount{Nonce: &n1}
		o.Add(OverrideAccount{})

		if o.Nonce == nil || uint64(*o.Nonce) != 1 {
			t.Fatalf("Nonce = %v, want 1", o.Nonce)
		}
	})

	t.Run("keeps other when receiver unset", func(t *testing.T) {
		n2 := hexutil.Uint64(2)
		o := OverrideAccount{}
		o.Add(OverrideAccount{Nonce: &n2})

		if o.Nonce == nil || uint64(*o.Nonce) != 2 {
			t.Fatalf("Nonce = %v, want 2", o.Nonce)
		}
	})
}

func TestOverrideAccount_Add_Code(t *testing.T) {
	codeA := hexutil.Bytes{0x01, 0x02}
	codeB := hexutil.Bytes{0x03, 0x04}

	t.Run("other wins when both set", func(t *testing.T) {
		o := OverrideAccount{Code: append(hexutil.Bytes{}, codeA...)}
		o.Add(OverrideAccount{Code: append(hexutil.Bytes{}, codeB...)})

		if !bytes.Equal(o.Code, codeB) {
			t.Fatalf("Code = %x, want %x", []byte(o.Code), []byte(codeB))
		}
	})

	t.Run("keeps receiver when other unset", func(t *testing.T) {
		o := OverrideAccount{Code: append(hexutil.Bytes{}, codeA...)}
		o.Add(OverrideAccount{})

		if !bytes.Equal(o.Code, codeA) {
			t.Fatalf("Code = %x, want %x", []byte(o.Code), []byte(codeA))
		}
	})

	t.Run("keeps other when receiver unset", func(t *testing.T) {
		o := OverrideAccount{}
		o.Add(OverrideAccount{Code: append(hexutil.Bytes{}, codeB...)})

		if !bytes.Equal(o.Code, codeB) {
			t.Fatalf("Code = %x, want %x", []byte(o.Code), []byte(codeB))
		}
	})
}

func TestOverrideAccount_Add_StateDiff(t *testing.T) {
	slotA := common.HexToHash("0x01")
	slotB := common.HexToHash("0x02")
	slotC := common.HexToHash("0x03")

	t.Run("sums single overlapping slot", func(t *testing.T) {
		o := OverrideAccount{
			StateDiff: map[common.Hash]common.Hash{
				slotA: common.BigToHash(big.NewInt(10)),
			},
		}
		o.Add(OverrideAccount{
			StateDiff: map[common.Hash]common.Hash{
				slotA: common.BigToHash(big.NewInt(5)),
			},
		})

		if got := new(big.Int).SetBytes(o.StateDiff[slotA].Bytes()); got.Int64() != 15 {
			t.Fatalf("StateDiff[slotA] = %s, want 15", got)
		}
	})

	t.Run("copies receiver-only slots while summing overlap", func(t *testing.T) {
		o := OverrideAccount{
			StateDiff: map[common.Hash]common.Hash{
				slotA: common.BigToHash(big.NewInt(10)),
				slotB: common.BigToHash(big.NewInt(20)),
			},
		}
		o.Add(OverrideAccount{
			StateDiff: map[common.Hash]common.Hash{
				slotA: common.BigToHash(big.NewInt(5)),
			},
		})

		if got := new(big.Int).SetBytes(o.StateDiff[slotA].Bytes()); got.Int64() != 15 {
			t.Fatalf("StateDiff[slotA] = %s, want 15", got)
		}
		if got := new(big.Int).SetBytes(o.StateDiff[slotB].Bytes()); got.Int64() != 20 {
			t.Fatalf("StateDiff[slotB] = %s, want 20", got)
		}
	})

	t.Run("unions other-only slots and copies receiver slots", func(t *testing.T) {
		o := OverrideAccount{
			StateDiff: map[common.Hash]common.Hash{
				slotA: common.BigToHash(big.NewInt(10)),
			},
		}
		o.Add(OverrideAccount{
			StateDiff: map[common.Hash]common.Hash{
				slotC: common.BigToHash(big.NewInt(30)),
			},
		})

		if got := new(big.Int).SetBytes(o.StateDiff[slotA].Bytes()); got.Int64() != 10 {
			t.Fatalf("StateDiff[slotA] = %s, want 10", got)
		}
		if got := new(big.Int).SetBytes(o.StateDiff[slotC].Bytes()); got.Int64() != 30 {
			t.Fatalf("StateDiff[slotC] = %s, want 30", got)
		}
	})

	t.Run("sums overlapping slots and unions unique slots", func(t *testing.T) {
		o := OverrideAccount{
			StateDiff: map[common.Hash]common.Hash{
				slotA: common.BigToHash(big.NewInt(10)),
				slotB: common.BigToHash(big.NewInt(20)),
			},
		}
		o.Add(OverrideAccount{
			StateDiff: map[common.Hash]common.Hash{
				slotA: common.BigToHash(big.NewInt(5)),
				slotC: common.BigToHash(big.NewInt(30)),
			},
		})

		if got := new(big.Int).SetBytes(o.StateDiff[slotA].Bytes()); got.Int64() != 15 {
			t.Fatalf("StateDiff[slotA] = %s, want 15", got)
		}
		if got := new(big.Int).SetBytes(o.StateDiff[slotB].Bytes()); got.Int64() != 20 {
			t.Fatalf("StateDiff[slotB] = %s, want 20", got)
		}
		if got := new(big.Int).SetBytes(o.StateDiff[slotC].Bytes()); got.Int64() != 30 {
			t.Fatalf("StateDiff[slotC] = %s, want 30", got)
		}
	})

	t.Run("keeps receiver slots when other has no StateDiff", func(t *testing.T) {
		o := OverrideAccount{
			StateDiff: map[common.Hash]common.Hash{
				slotA: common.BigToHash(big.NewInt(10)),
			},
		}
		o.Add(OverrideAccount{})

		if got := new(big.Int).SetBytes(o.StateDiff[slotA].Bytes()); got.Int64() != 10 {
			t.Fatalf("StateDiff[slotA] = %s, want 10", got)
		}
	})
}

func TestOverrideAccount_Add_State(t *testing.T) {
	slotA := common.HexToHash("0x0a")
	slotB := common.HexToHash("0x0b")
	slotC := common.HexToHash("0x0c")

	t.Run("sums single overlapping slot", func(t *testing.T) {
		o := OverrideAccount{
			State: map[common.Hash]common.Hash{
				slotA: common.BigToHash(big.NewInt(7)),
			},
		}
		o.Add(OverrideAccount{
			State: map[common.Hash]common.Hash{
				slotA: common.BigToHash(big.NewInt(3)),
			},
		})

		if got := new(big.Int).SetBytes(o.State[slotA].Bytes()); got.Int64() != 10 {
			t.Fatalf("State[slotA] = %s, want 10", got)
		}
	})

	t.Run("copies receiver-only slots while summing overlap", func(t *testing.T) {
		o := OverrideAccount{
			State: map[common.Hash]common.Hash{
				slotA: common.BigToHash(big.NewInt(7)),
				slotB: common.BigToHash(big.NewInt(9)),
			},
		}
		o.Add(OverrideAccount{
			State: map[common.Hash]common.Hash{
				slotA: common.BigToHash(big.NewInt(3)),
			},
		})

		if got := new(big.Int).SetBytes(o.State[slotA].Bytes()); got.Int64() != 10 {
			t.Fatalf("State[slotA] = %s, want 10", got)
		}
		if got := new(big.Int).SetBytes(o.State[slotB].Bytes()); got.Int64() != 9 {
			t.Fatalf("State[slotB] = %s, want 9", got)
		}
	})

	t.Run("sums overlapping slots and unions unique slots", func(t *testing.T) {
		o := OverrideAccount{
			State: map[common.Hash]common.Hash{
				slotA: common.BigToHash(big.NewInt(7)),
				slotB: common.BigToHash(big.NewInt(9)),
			},
		}
		o.Add(OverrideAccount{
			State: map[common.Hash]common.Hash{
				slotA: common.BigToHash(big.NewInt(3)),
				slotC: common.BigToHash(big.NewInt(11)),
			},
		})

		if got := new(big.Int).SetBytes(o.State[slotA].Bytes()); got.Int64() != 10 {
			t.Fatalf("State[slotA] = %s, want 10", got)
		}
		if got := new(big.Int).SetBytes(o.State[slotB].Bytes()); got.Int64() != 9 {
			t.Fatalf("State[slotB] = %s, want 9", got)
		}
		if got := new(big.Int).SetBytes(o.State[slotC].Bytes()); got.Int64() != 11 {
			t.Fatalf("State[slotC] = %s, want 11", got)
		}
	})

	t.Run("keeps receiver slots when other has no State", func(t *testing.T) {
		o := OverrideAccount{
			State: map[common.Hash]common.Hash{
				slotA: common.BigToHash(big.NewInt(7)),
			},
		}
		o.Add(OverrideAccount{})

		if got := new(big.Int).SetBytes(o.State[slotA].Bytes()); got.Int64() != 7 {
			t.Fatalf("State[slotA] = %s, want 7", got)
		}
	})
}

func TestStateOverride_Add(t *testing.T) {
	addr1 := common.HexToAddress("0x1111111111111111111111111111111111111111")
	addr2 := common.HexToAddress("0x2222222222222222222222222222222222222222")

	t.Run("inserts new address", func(t *testing.T) {
		s := StateOverride{}
		s.Add(addr1, OverrideAccount{Balance: balanceHex(100)})

		got, ok := s[addr1]
		if !ok {
			t.Fatal("expected address to be present")
		}
		if got.Balance == nil || got.Balance.ToInt().Int64() != 100 {
			t.Fatalf("Balance = %v, want 100", got.Balance)
		}
	})

	t.Run("merges into existing address", func(t *testing.T) {
		s := StateOverride{
			addr1: {Balance: balanceHex(100)},
		}
		s.Add(addr1, OverrideAccount{Balance: balanceHex(25)})

		got := s[addr1]
		if got.Balance == nil || got.Balance.ToInt().Int64() != 125 {
			t.Fatalf("Balance = %v, want 125", got.Balance)
		}
	})

	t.Run("keeps distinct addresses independent", func(t *testing.T) {
		s := StateOverride{}
		s.Add(addr1, OverrideAccount{Balance: balanceHex(100)})
		s.Add(addr2, OverrideAccount{Balance: balanceHex(200)})

		if s[addr1].Balance.ToInt().Int64() != 100 {
			t.Fatalf("addr1 Balance = %v, want 100", s[addr1].Balance)
		}
		if s[addr2].Balance.ToInt().Int64() != 200 {
			t.Fatalf("addr2 Balance = %v, want 200", s[addr2].Balance)
		}
	})
}

func balanceHex(v int64) *hexutil.Big {
	b := hexutil.Big(*big.NewInt(v))
	return &b
}
