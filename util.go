package main

import (
	"strconv"
	"strings"
)

func notPrimary() bool {
	return !isPrimary
}

func stringToByteSlice(s string) []byte {
	ss := strings.Split(s, ":")
	data := make([]byte, len(ss))
	for i, b := range ss {
		d, _ := strconv.ParseUint(b, 16, 8)
		data[i] = byte(d)
	}
	return data
}

func byteSliceToString(bytes []byte) string {
	var builder strings.Builder
	for _, b := range bytes {
		builder.WriteString(strconv.FormatUint(uint64(b), 16))
		builder.WriteByte(byte(':'))
	}
	res := builder.String()
	res = res[:len(res)-1]
	return res
}
