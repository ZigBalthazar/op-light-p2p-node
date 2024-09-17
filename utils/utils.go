package utils

import (
	"fmt"
	"math/big"
	"os"
)

func HexToDecimal(hexStr string) (*big.Int, error) {
	decimal := new(big.Int)
	_, success := decimal.SetString(hexStr[2:], 16)
	if !success {
		return nil, fmt.Errorf("failed to parse hexadecimal number")
	}
	return decimal, nil
}

func EnvVariable(key string) string {
	return os.Getenv(key)
}
