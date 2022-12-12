package utils

import "encoding/binary"

func Int64ToByteArray(val int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(val))
	return b
}

func ByteArrayToInt64(val []byte) int64 {
	return int64(binary.BigEndian.Uint64(val))
}
