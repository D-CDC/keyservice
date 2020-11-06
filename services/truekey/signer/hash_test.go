package signer

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"ethereum/keyservice/common"
	"ethereum/keyservice/common/hexutil"
	"ethereum/keyservice/core/types"
	"ethereum/keyservice/crypto"
	"ethereum/keyservice/crypto/ecies"
	"ethereum/keyservice/log"
	"ethereum/keyservice/rlp"
	clitype "ethereum/keyservice/services/truekey/types"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"testing"
)

func TestHash(t *testing.T) {
	priv1, err := crypto.HexToECDSA("0b02806b49d538c9f77843c4d0b6296ef087722a927a23ac88ac8bc95607c4bc")
	if err != nil {
		fmt.Println("err", err)
	}
	address := common.HexToAddress("0xaB4D1a46F0F9331201042C359f00C81537741673")
	// Create the transaction.
	tx := types.NewTransaction(0, address, big.NewInt(1), 30000, big.NewInt(1000000000), nil)
	tx, _ = types.SignTx(tx, types.NewTIP1Signer(new(big.Int).SetUint64(18928)), priv1)
	v, r, s := tx.RawSignatureValues()
	fmt.Println("v ", v, " r ", r, " s ", s)
	fmt.Println("r ", hexutil.Encode(r.Bytes()), " s ", hexutil.Encode(s.Bytes()))
	fmt.Println("r ", r.Bytes(), " s ", s.Bytes())
	fmt.Println("pub ", hexutil.Encode(crypto.FromECDSAPub(&priv1.PublicKey)))
	data := "045682e3ae7a17daa25982b2fdbb9a5e7ef388a4818dfe93a3d05813f081fd865b8ccfecc8880f4b117173a12bdf43a5e012723a4cd12877c800ab8825e363a1d4ad08f3844395a340c726c3bb40a75d30a975b2890b2542539a9538fde2fbca72ee052bd8db5cef91e58f92349d993334c493a54ccd44cef89f9005f9cff8dbed0cbccabff0e7a44a87d5ac421ad9afc372a8305e4567cc52945971c337dd1bf21bcf88a0accfc9730d5eebeffe91ee7af905e6206641b39b941b0a67204f4fc920f94dc5f870d135597db3db950a47c7558cc8377861496aa6100f883d0f68ad733795f817dcb022eeca0d3cc4a7aef58700ea827576bba969abc4332a287361076fa0acbeaadf7979a7af7d68316ddc2704e1b04a705a6bf017e88e3b65e70726eb196787bb494b12cb7154b84954d70e5e192f2f5c124803be9db7a86cbd6fc720eae2fa06b5f805458b015a913a4432eefe1bf931b032b1ad4c64b21a26f1f4c0612b9743742aecfe600b465de67d9c0bb7ee27e671ad1be125da94948186"
	input, err := hex.DecodeString(data)
	if err != nil {
		fmt.Println("input err", err)
	}
	priKey := ecies.ImportECDSA(priv1)
	decryptMessage, err := priKey.Decrypt(input, nil, nil)
	if err != nil {
		fmt.Println("decryptMessage input err", err)
	}

	enc := base64.StdEncoding
	buf := make([]byte, enc.DecodedLen(len(decryptMessage)))
	i, err := enc.Decode(buf, decryptMessage)
	fmt.Println("i", i, err, err, "buf ", hex.EncodeToString(buf), " decryptMessage ", hex.EncodeToString(decryptMessage))

	enc = base64.StdEncoding
	buf1 := make([]byte, enc.EncodedLen(len(buf)))
	enc.Encode(buf1, buf)
	fmt.Println("input ", hex.EncodeToString(buf1))

	query := new(clitype.ClentQuest)
	if err := rlp.DecodeBytes(buf, query); err != nil {
		fmt.Println("Sign Hash failed to decode decrypt message", "err", err)
	}

	fmt.Println("decryptMessage", hex.EncodeToString(decryptMessage))

	fmt.Println("query", query.Data, "query", query.Pub)

	data = "0xc2702f8379e539daaf86396b8962c382ccdede31782d6d843883585a815817e0"
	_, err = hex.DecodeString(data)
	if err != nil {
		log.Error("DecodeString error: ", "err", err)
	}
	pub := "0x04d68d95f60c315fe2bd674db08b468eeee87ac55ec4f1713e5cba738b8a2a26895e5a97eef6d7d7f968eb2f3a363ea7a96ba08fccefffd2bb1375ae7128a87bcc"
	_, err = hexutil.Decode(pub)
	if err != nil {
		log.Error("Decode pub error: ", "err", err)
	}
	val := clitype.ClentQuest{Data: data, Pub: pub}
	resultByte, err := rlp.EncodeToBytes(val)
	if err != nil {
		log.Error("EncodeToBytes error: ", "err", err)
	}
	fmt.Println(" rlp data ", hex.EncodeToString(resultByte))

	fmt.Println(" base64 data ", hex.EncodeToString(buf))
	encryptMessageInfo, err := ecies.Encrypt(rand.Reader, ecies.ImportECDSAPublic(&priv1.PublicKey), input, nil, nil)
	if err != nil {
		log.Error("publickey encrypt result error ", "publickey", common.Bytes2Hex(crypto.FromECDSAPub(&priv1.PublicKey)), "err", err)
	}
	fmt.Println(" encryptMessageInfo ", hex.EncodeToString(encryptMessageInfo))
	//cryMessage := &clitype.EncryptMessage{
	//	CreatedAt: uint64(time.Now().Unix()),
	//}

	data = "5e64a6fa1e829fb034d019626c2fa6b730c0a95f1efd65d54fc10f74002f7c2e1184f8f28786e94ce0e078d2638cd8b11fd26d341a2a1cdf2e645b1a8f95472701"
	input, err = hex.DecodeString(data)
	if err != nil {
		fmt.Println("input err", err)
	}
	enc = base64.RawURLEncoding
	buf = make([]byte, enc.EncodedLen(len(input)))
	enc.Encode(buf, input)
	fmt.Println("buf ", string(buf))
	fmt.Println("", hexutil.Encode(buf), "old ", hexutil.Encode(input))

	priv1, err = crypto.HexToECDSA("7eab39ba2e3f01ea7eeccfa7cf66bf6fada5352687c38c0d5b2b5cc008f288d1")
	if err != nil {
		fmt.Println("err", err)
	}
	input, err = hex.DecodeString("04d7a68985f39d7d6580c3bafd4b192923273ecb4f187f47f5b15dc60e0d3c29f6abd464232fc3151664ca73d7dd5e7180ac3165997cd4d3e5c12bce29162e1b61908003451b7583779e8a284ecbd8e28d5c4a54320eceb63fb5c5a01f03b7b84f3f1ece50f91f6ba06657edd92f0e5417aef792a3ac0508372fc67fd1a0badae6c5a2269343a4e65b9ea4ab7f30624e85e6f564fcd2d2b79226aedd1de2cc6c4c1606a79c308d2b28be6a9e6de0419d77c7b4d0e44a2d12cd91f4903db6a0a8e5e1400ccf89e38d9a")
	if err != nil {
		fmt.Println("input err", err)
	}
	priKey = ecies.ImportECDSA(priv1)
	decryptMessage, err = priKey.Decrypt(input, nil, nil)
	if err != nil {
		fmt.Println("decryptMessage input err", err)
	}
	fmt.Println("Decrypt input ", input)
	fmt.Println("decryptMessage", hex.EncodeToString(decryptMessage))
	enc = base64.StdEncoding
	buf = make([]byte, enc.DecodedLen(len(decryptMessage)))
	i, err = enc.Decode(buf, decryptMessage)
	fmt.Println("i", i, err, err, "buf ", hex.EncodeToString(buf), " decryptMessage ", hex.EncodeToString(decryptMessage))

}

func TestGeneratePub(t *testing.T) {
	key, _ := crypto.GenerateKey()
	pubdata := crypto.FromECDSAPub(&key.PublicKey)
	fmt.Println("hex", hex.EncodeToString(pubdata))
	addr1 := common.HexToAddress("0x31941333f5a6503ea0f362a2db58d1bd9ee0d33c")
	addr2 := common.HexToAddress("0x31941333F5a6503ea0f362A2DB58d1bD9Ee0d33c")

	fmt.Println(addr1 == addr2)
	contracts := make(map[common.Address]string)
	contracts[addr1] = "1"
	if v, ok := contracts[addr2]; ok {
		fmt.Println(" v ", v)
	}

}

func TestSignature(t *testing.T) {
	sign, err := hex.DecodeString("40984ad9f4cc809931fe6d787467e41abdce3b5a93c8ecfa35c8f831d9aa80561a079b9774066b4e77ff8fcdab51057457ec2477bd9ebdd8ff04c6667046bc8e1b")
	if err != nil {
		fmt.Println("data err", err)
	}
	hash := common.HexToHash("0xac0ef8a6fa119bd4d6e7eb20880a398ead6a646127938da5d85647affbff0004")
	pubKey, err := crypto.SigToPub(hash[:], sign)
	if err != nil {
		fmt.Println("pubKey err", err)
	}
	fmt.Println("pub ", hex.EncodeToString(crypto.FromECDSAPub(pubKey)))
	priv1, err := crypto.HexToECDSA("0b02806b49d538c9f77843c4d0b6296ef087722a927a23ac88ac8bc95607c4bc")
	if err != nil {
		fmt.Println("err", err)
	}
	// Create the transaction.
	fmt.Println("pub ", hexutil.Encode(crypto.FromECDSAPub(&priv1.PublicKey)))
}

func TestParseDerivationPath(t *testing.T) {
	index := 1000
	arrs := strings.Split(DefaultBaseDerivationPath, "/")
	arrs[3] = strconv.FormatInt(int64(index), 10) + "'"
	fmt.Println(arrs)
	subDerivationPath := strings.Join(arrs, "/")
	pathIndex := subDerivationPath + fmt.Sprintf("%d", index)
	fmt.Println(pathIndex)
}

type Phone struct {
	Phone int `json:"userId"`
}

func TestParse(t *testing.T) {
	arrs := []string{"1", "2", "3"}
	fmt.Println(strings.Join(arrs, ","))
	path, err := GetDerivationPath(9999)
	fmt.Println(path, " ", err)
	jsonStr := "{\"userId\":9999,\"gas_price\":9999,\"gas_limit\":9999,\"nonce\":9999,\"chain_id\":9999,\"data\":\"0x00\"}"
	var phoneNumber Phone
	err = json.Unmarshal([]byte(jsonStr), &phoneNumber)
	if err != nil {
		fmt.Println(err, " ", phoneNumber.Phone)
	}

	var tx clitype.SignTx
	err = json.Unmarshal([]byte(jsonStr), &tx)
	if err != nil {
		fmt.Println(err)
	}
}
func TestPathParse(t *testing.T) {
	arrs := []string{"1", "2", "3"}
	fmt.Println(strings.Join(arrs, ","))
	path, err := GetDerivationPath(4294967290)
	if err != nil {
		fmt.Println(err, " ", path)
	}
}
