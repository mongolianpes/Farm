package crypto

import (
	"fmt"
	"testing"
)

func TestRandHex32(t *testing.T) {
	id, err := RandHex32()
	if err != nil {
		t.Error(err)
	}
	fmt.Println("len: ", len(id))
	fmt.Println(id)
}
