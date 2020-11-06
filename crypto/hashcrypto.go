package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"errors"
	"fmt"
)

const VECVI = "1234560405060708"

//plainText fill
func Padding(plainText []byte, blockSize int) []byte {
	//cal len
	n := blockSize - len(plainText)%blockSize
	//fil n
	temp := bytes.Repeat([]byte{byte(n)}, n)
	plainText = append(plainText, temp...)
	return plainText
}

//delete fill
func UnPadding(cipherText []byte) ([]byte, error) {
	//get last one byte
	end := cipherText[len(cipherText)-1]
	//delete fil
	if len(cipherText)-int(end) < 0 {
		return nil, errors.New(fmt.Sprintf("len(cipherText) = %d end = %d", len(cipherText), end))
	}
	cipherText = cipherText[:len(cipherText)-int(end)]
	return cipherText, nil
}

//AEC Encrypt（CBC mode）
func AESCbCEncrypt(plainText []byte, key []byte) ([]byte, error) {
	//return AES Block interface
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	//fill
	plainText = Padding(plainText, block.BlockSize())
	//assign vector vi,len and accordance block
	iv := []byte(VECVI)
	blockMode := cipher.NewCBCEncrypter(block, iv)
	cipherText := make([]byte, len(plainText))
	blockMode.CryptBlocks(cipherText, plainText)
	return cipherText, nil
}

func AESCbCDecrypt(cipherText []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	iv := []byte(VECVI)
	blockMode := cipher.NewCBCDecrypter(block, iv)
	plainText := make([]byte, len(cipherText))
	blockMode.CryptBlocks(plainText, cipherText)
	plainText, err = UnPadding(plainText)
	if err != nil {
		return nil, err
	}
	return plainText, nil
}
