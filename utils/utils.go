package utils

import (
	"encoding/json"
	"fmt"
	"math/big"
)

func HexToDecimal(hexStr string) (*big.Int, error) {
	decimal := new(big.Int)
	_, success := decimal.SetString(hexStr[2:], 16)
	if !success {
		return nil, fmt.Errorf("failed to parse hexadecimal number")
	}
	return decimal, nil
}

func StructToMap(data interface{}) (map[string]interface{}, error) {
    var result map[string]interface{}
    bytes, err := json.Marshal(data)
    if err != nil {
        return nil, err
    }
    err = json.Unmarshal(bytes, &result)
    if err != nil {
        return nil, err
    }
    return result, nil
}


func ToJsonString(data interface{}) (string, error) {
	jsonString, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return "", err
	}
	return string(jsonString), nil
}