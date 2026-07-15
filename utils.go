package multicall

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
)

// Add0xPrefix adds 0x hex prefix to a string, if needed.
func Add0xPrefix(s string) string {
	if strings.HasPrefix(s, "0x") {
		return s
	}
	return "0x" + s
}

func mergeStorageMaps(dst, src map[common.Hash]common.Hash) map[common.Hash]common.Hash {
	if len(src) == 0 {
		return dst
	}
	if dst == nil {
		dst = make(map[common.Hash]common.Hash, len(src))
	}
	for k, v := range src {
		if existing, ok := dst[k]; ok {
			summed := new(big.Int).Add(
				new(big.Int).SetBytes(existing.Bytes()),
				new(big.Int).SetBytes(v.Bytes()),
			)
			dst[k] = common.BigToHash(summed)
			continue
		}
		dst[k] = v
	}
	return dst
}
