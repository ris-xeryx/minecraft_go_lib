package mcgo

import (
	"crypto/md5"
	"fmt"
)

// md5Bytes son los bytes del MD5.
func md5Bytes(data []byte) []byte {
	h := md5.Sum(data)
	return h[:]
}

// formatUUID formatea 16 bytes como UUID v3 (a la Java): xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
func formatUUID(b []byte) string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}
