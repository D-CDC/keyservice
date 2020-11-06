package types

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestCheckIp(t *testing.T) {
	ips := []string{"*.*.*.*", "*.*.*", "127.2.3.4", "124.3.*.5", "123.3.4.*", "123.*.1.4", "1234.3.4.*", "*.*.*.**", "91.1.2.*"}
	ip := "121"
	fmt.Println(CheckIp(ips), " len", len(ip))
}

func TestContainIp(t *testing.T) {
	ip1 := "127.0.0.1"
	src := "127.0.1.*"
	fmt.Println(ContainIp(ip1, src), " len", len(src))
}

func TestRandString(t *testing.T) {
	fmt.Println(hex.EncodeToString(RandString(16)))
}
