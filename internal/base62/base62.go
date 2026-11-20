package base62

import "math/big"

func EncodeID(id int64) string {
	return big.NewInt(id).Text(62)
}
