package crypto

import (
	"encoding/hex"
	"fmt"
	"testing"
)

func TestAESCbCEncrypt(t *testing.T) {
	data, err := AESCbCEncrypt([]byte("abcd"), []byte("YNMPWZMUNAGMSGPW"))
	if err != nil {
		fmt.Println("err", err)
	}
	fmt.Println("", hex.EncodeToString(data))
	data, err = AESCbCDecrypt(data, []byte("YNMPWZMUNAGMSGPW"))
	if err != nil {
		fmt.Println("AESCbCDecrypt err", err)
	}
	fmt.Println("", string(data))

	signH := "8e402e8440bfed8e2591ac03029e5a430fdd38097be799cc4e77afd079fb12360dabb9c4cf431191fb989a8b696feb45e984343f62f6b26b161bf8a05bc69952c0f30ee06a8dcddadcded7329cb2dd29"
	sign, err := hex.DecodeString(signH)
	if err != nil {
		fmt.Println("AESCbCDecrypt key err", err)
	}
	key, err := hex.DecodeString("47a569d4e76429a1368797f669a7f2bd")
	if err != nil {
		fmt.Println("AESCbCDecrypt key err", err)
	}
	data, err = AESCbCDecrypt(sign, key)
	if err != nil {
		fmt.Println("AESCbCDecrypt err", err)
	}
	fmt.Println("", string(data), " data ", hex.EncodeToString(data))

	priv1, err := HexToECDSA("93fccdf766fb1bf22f307035f9e4d823ed92a386bd62cefda6583f7c9aabec81")
	if err != nil {
		fmt.Println("err", err)
	}

	data, err = hex.DecodeString(string(data))
	if err != nil {
		fmt.Printf("%v: %v \n", "Sign decode string error", err)
	}
	fmt.Println("hash ", hex.EncodeToString(data), " data ", string(data))
	result, err := Sign(data, priv1)
	if err != nil {
		fmt.Printf("%v: %v \n", "Sign decode string error", err)
	}
	//c2702f8379e539daaf86396b8962c382ccdede31782d6d843883585a815817e0
	fmt.Println("result ", hex.EncodeToString(result))

	data, err = AESCbCEncrypt([]byte(hex.EncodeToString(result)), key)
	if err != nil {
		fmt.Println("err", err)
	}
	fmt.Println("", hex.EncodeToString(data))
}
