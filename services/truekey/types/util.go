package types

import (
	"errors"
	"ethereum/keyservice/common"
	"ethereum/keyservice/rlp"
	"golang.org/x/crypto/sha3"
	"math/rand"
	"net"
	"strings"
	"time"
)

var (
	ErrDappNotRegister  = errors.New("dapp not exist,please call RegisterDapp")
	ErrRootError        = errors.New("root id error")
	ErrRootNotServer    = errors.New("root keystore not server")
	ErrAdminError       = errors.New("admin not exist in server")
	ErrAdminSignError   = errors.New("admin sign error")
	ErrDappAlready      = errors.New("dapp already exist")
	ErrAccountNotExist  = errors.New("account not exist")
	ErrChildNotExist    = errors.New("child id not exist")
	ErrAccountLock      = errors.New("account lock")
	ErrDappIP           = errors.New("ip not in dapp whitelist")
	ErrDappPubError     = errors.New("dapp pub error")
	ErrEncryptDataError = errors.New("encrypt result data error")
	ErrSignTxError      = errors.New("sign tx error")
	ErrPaymentError     = errors.New("not payment error")
	ErrCreateTxError    = errors.New("create tx error")
	ErrPhoneError       = errors.New("phone number error")
	ErrPhoneNumberError = errors.New("phone number spilt error")
)

func CheckIp(ips []string) []string {
	var correctIps []string
	for _, ip := range ips {
		if net.ParseIP(ip) != nil {
			correctIps = append(correctIps, ip)
		} else {
			if strings.ContainsAny(ip, "*") {
				arr := strings.Split(ip, ".")
				if len(arr) != 4 {
					continue
				}
				k := 0
				for m, v := range arr {
					//*.*.*.*
					if len(v) == 1 && v == "*" {
						//1.2.*.4
						if m != len(arr)-1 && arr[m+1] != "*" {
							break
						}
						k = k + 1
						continue
					}

					if len(v) > 3 {
						break
					}
					n := 0
					find := false
					for i := 0; i < len(v); i++ {
						if '0' > v[i] && v[i] > '9' {
							find = true
							break
						}
						n = n*10 + int(v[i]-'0')
						if n >= 256 {
							find = true
							break
						}
					}
					if find {
						break
					}
					k++
				}
				if k == 4 {
					correctIps = append(correctIps, ip)
				}
			}
		}
	}
	return correctIps
}

func ContainIp(s, src string) bool {
	if s == src {
		return true
	}
	arr1 := strings.Split(s, ".")
	arr2 := strings.Split(src, ".")

	for i, v := range arr2 {
		if arr1[i] != v && v != "*" {
			return false
		}
	}
	return true
}

func rlpHash(x interface{}) (h common.Hash) {
	hw := sha3.NewLegacyKeccak256()
	rlp.Encode(hw, x)
	hw.Sum(h[:0])
	return h
}

func RandString(len int) []byte {
	r := rand.New(rand.NewSource(time.Now().Unix()))
	bytes := make([]byte, len)
	for i := 0; i < len; i++ {
		b := r.Intn(256)
		bytes[i] = byte(b)
	}
	return bytes
}
