package hash

import "hash/crc32"

func Hash(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
}
