package hash

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

func Hash(input ...string) string {
	elements := strings.Join(input, "_")
	hash := sha256.Sum256([]byte(elements))
	return hex.EncodeToString(hash[:16])
}
