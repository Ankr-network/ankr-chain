package common

import (
	"encoding/binary"
	"encoding/hex"
	"testing"
)

func TestAppHashParse(t *testing.T) {
	appHashBytes1, _ := hex.DecodeString("E69A420000000000")
	appHashBytes2, _ := hex.DecodeString("E49A420000000000")

	h1, _ := binary.Varint(appHashBytes1)
	h2, _ := binary.Varint(appHashBytes2)

	t.Logf("h1=%d, h2=%d\n", h1, h2)
}
