// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package types

import (
	"container/heap"
	"errors"
	"io"
	"math/big"
	"sync/atomic"

	"fmt"
	"strconv"

	"ethereum/keyservice/common"
	"ethereum/keyservice/common/hexutil"
	"ethereum/keyservice/crypto"
	"ethereum/keyservice/rlp"
)

//go:generate gencodec -type txdata -field-override txdataMarshaling -out gen_tx_json.go

var (
	ErrInvalidSig = errors.New("invalid transaction v, r, s values")
)

type Transaction struct {
	data txdata
	// caches
	hash    atomic.Value
	size    atomic.Value
	from    atomic.Value
	payment atomic.Value
}

type RawTransaction struct {
	data raw_txdata
	// caches
	hash atomic.Value
	size atomic.Value
	from atomic.Value
}

type txdata struct {
	AccountNonce uint64          `json:"nonce"    gencodec:"required"`
	Price        *big.Int        `json:"gasPrice" gencodec:"required"`
	GasLimit     uint64          `json:"gas"      gencodec:"required"`
	Recipient    *common.Address `json:"to"       rlp:"nil"` // nil means contract creation
	Amount       *big.Int        `json:"value"    gencodec:"required"`
	Payload      []byte          `json:"input"    gencodec:"required"`
	Payer        *common.Address `json:"payer"    rlp:"nil"`
	Fee          *big.Int        `json:"fee"   rlp:"nil"`

	// Signature values
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`

	// Paied Signature values
	PV *big.Int `json:"pv" rlp:"nil"` // nil means donnot have payment
	PR *big.Int `json:"pr" rlp:"nil"`
	PS *big.Int `json:"ps" rlp:"nil"`

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"hash" rlp:"-"`
}

type raw_txdata struct {
	AccountNonce uint64          `json:"nonce"    gencodec:"required"`
	Price        *big.Int        `json:"gasPrice" gencodec:"required"`
	GasLimit     uint64          `json:"gas"      gencodec:"required"`
	Recipient    *common.Address `json:"to"       rlp:"nil"` // nil means contract creation
	Amount       *big.Int        `json:"value"    gencodec:"required"`
	Payload      []byte          `json:"input"    gencodec:"required"`

	// Signature values
	V *big.Int `json:"v" gencodec:"required"`
	R *big.Int `json:"r" gencodec:"required"`
	S *big.Int `json:"s" gencodec:"required"`

	// This is only used when marshaling to JSON.
	Hash *common.Hash `json:"hash" rlp:"-"`
}

func (rawTransaction *RawTransaction) ConvertTransaction() *Transaction {
	cpy := &RawTransaction{data: rawTransaction.data}
	cpy_data := cpy.data

	tx := new(Transaction)
	//fmt.Println("data.Recipient", cpy_data.Recipient)
	if cpy_data.Recipient == nil {
		tx = NewContractCreation(cpy_data.AccountNonce, cpy_data.Amount, cpy_data.GasLimit, cpy_data.Price, cpy_data.Payload)
	} else {
		tx = NewTransaction(cpy_data.AccountNonce, *cpy_data.Recipient, cpy_data.Amount, cpy_data.GasLimit, cpy_data.Price, cpy_data.Payload)
	}
	tx.data.V = cpy_data.V
	tx.data.R = cpy_data.R
	tx.data.S = cpy_data.S
	return tx
}

func (tx *Transaction) ConvertRawTransaction() *RawTransaction {
	cpy := &Transaction{data: tx.data}
	cpy_data := cpy.data

	raw_tx := new(RawTransaction)
	if cpy_data.Recipient == nil {
		raw_tx = NewRawTransactionContract(cpy_data.AccountNonce, cpy_data.Amount, cpy_data.GasLimit, cpy_data.Price, cpy_data.Payload)
	} else {
		raw_tx = NewRawTransaction(cpy_data.AccountNonce, *cpy_data.Recipient, cpy_data.Amount, cpy_data.GasLimit, cpy_data.Price, cpy_data.Payload)
	}
	raw_tx.data.V = cpy_data.V
	raw_tx.data.R = cpy_data.R
	raw_tx.data.S = cpy_data.S
	return raw_tx
}

type txdataMarshaling struct {
	AccountNonce hexutil.Uint64
	Price        *hexutil.Big
	GasLimit     hexutil.Uint64
	Amount       *hexutil.Big
	Payload      hexutil.Bytes
	V            *hexutil.Big
	R            *hexutil.Big
	S            *hexutil.Big
}

func NewTransaction(nonce uint64, to common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	return NewTransaction_Payment(nonce, to, amount, nil, gasLimit, gasPrice, data, common.Address{})
}

func NewTransaction_Payment(nonce uint64, to common.Address, amount *big.Int, fee *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, payer common.Address) *Transaction {
	if payer == (common.Address{}) {
		return newTransaction(nonce, &to, nil, amount, fee, gasLimit, gasPrice, data)
	}
	return newTransaction(nonce, &to, &payer, amount, fee, gasLimit, gasPrice, data)
}

func NewContractCreation(nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	return NewContractCreation_Payment(nonce, amount, nil, gasLimit, gasPrice, data, common.Address{})
}

func NewContractCreation_Payment(nonce uint64, amount *big.Int, fee *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, payer common.Address) *Transaction {
	if payer == (common.Address{}) {
		return newTransaction(nonce, nil, nil, amount, fee, gasLimit, gasPrice, data)
	}
	return newTransaction(nonce, nil, &payer, amount, fee, gasLimit, gasPrice, data)
}

func newTransaction(nonce uint64, to *common.Address, payer *common.Address, amount *big.Int, fee *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *Transaction {
	if len(data) > 0 {
		data = common.CopyBytes(data)
	}
	d := txdata{
		AccountNonce: nonce,
		Recipient:    to,
		Payer:        payer,
		Payload:      data,
		Amount:       new(big.Int),
		GasLimit:     gasLimit,
		Price:        new(big.Int),
		V:            new(big.Int),
		R:            new(big.Int),
		S:            new(big.Int),
		PV:           new(big.Int),
		PR:           new(big.Int),
		PS:           new(big.Int),
	}
	if amount != nil {
		d.Amount.Set(amount)
	}
	if fee != nil {
		d.Fee = new(big.Int)
		d.Fee.Set(fee)
	}
	if gasPrice != nil {
		d.Price.Set(gasPrice)
	}

	return &Transaction{data: d}
}

func NewRawTransaction(nonce uint64, to common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *RawTransaction {
	return newRawTransaction(nonce, &to, amount, gasLimit, gasPrice, data)
}

func NewRawTransactionContract(nonce uint64, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *RawTransaction {
	return newRawTransaction(nonce, nil, amount, gasLimit, gasPrice, data)
}

func newRawTransaction(nonce uint64, to *common.Address, amount *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte) *RawTransaction {
	if len(data) > 0 {
		data = common.CopyBytes(data)
	}
	d := raw_txdata{
		AccountNonce: nonce,
		Recipient:    to,
		Payload:      data,
		Amount:       new(big.Int),
		GasLimit:     gasLimit,
		Price:        new(big.Int),
		V:            new(big.Int),
		R:            new(big.Int),
		S:            new(big.Int),
	}
	if amount != nil {
		d.Amount.Set(amount)
	}
	if gasPrice != nil {
		d.Price.Set(gasPrice)
	}
	return &RawTransaction{data: d}
}

// ChainId returns which chain id this transaction was signed for (if at all)
func (tx *Transaction) ChainId() *big.Int {
	return deriveChainId(tx.data.V)
}

// Protected returns whether the transaction is protected from replay protection.
func (tx *Transaction) Protected() bool {
	return isProtectedV(tx.data.V)
}

func (tx *Transaction) Protected_Payment() bool {
	return isProtectedV(tx.data.PV)
}

func isProtectedV(V *big.Int) bool {
	if V.BitLen() <= 8 {
		v := V.Uint64()
		return v != 27 && v != 28
	}
	// anything not 27 or 28 are considered unprotected
	return true
}

// EncodeRLP implements rlp.Encoder
func (tx *Transaction) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &tx.data)
}

// DecodeRLP implements rlp.Decoder
func (tx *Transaction) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()
	err := s.Decode(&tx.data)
	if err == nil {
		tx.size.Store(common.StorageSize(rlp.ListSize(size)))
	}

	return err
}

// MarshalJSON encodes the web3 RPC transaction format.
func (tx *Transaction) MarshalJSON() ([]byte, error) {
	hash := tx.Hash()
	data := tx.data
	data.Hash = &hash
	return data.MarshalJSON()
}

func (tx *Transaction) Info() string {
	str := ""
	payer := ""
	recipient := ""
	fee := ""
	payload := ""
	if tx.data.Fee != nil {
		fee = strconv.Itoa(int(tx.data.Fee.Int64()))
	}
	if tx.data.Payer != nil {
		payer = common.Bytes2Hex(tx.data.Payer[:])
	}
	if tx.data.Recipient != nil {
		recipient = common.Bytes2Hex(tx.data.Recipient[:])
	}
	if tx.data.Payload != nil {
		payload = common.Bytes2Hex(tx.data.Payload[:])
	}

	str += fmt.Sprintf("nonce=%v,price=%v, gaslimit=%v,Recipient=%v,Amount=%v,Payload=%v,chainId=%v,fee=%v,payment=%v, v=%v,r=%v,s=%v,",
		tx.data.AccountNonce, tx.data.Price, tx.data.GasLimit, recipient, tx.data.Amount, payload, tx.ChainId(),
		fee, payer, tx.data.V, tx.data.R, tx.data.S)
	return str
}

func (tx *RawTransaction) Info() string {
	recipient := ""
	if tx.data.Recipient != nil {
		recipient = tx.data.Recipient.String()
	}
	str := ""
	if tx != nil {
		str += fmt.Sprintf("nonce=%v,price=%v gaslimit=%v,Recipient=%v,Amount=%v,Payload=%v v=%v,r=%v,s=%v,",
			tx.data.AccountNonce, tx.data.Price, tx.data.GasLimit, recipient, tx.data.Amount, tx.data.Payload,
			tx.data.V, tx.data.R, tx.data.S)
	}
	return str
}

// EncodeRLP implements rlp.Encoder
func (tx *RawTransaction) EncodeRLP(w io.Writer) error {
	return rlp.Encode(w, &tx.data)
}

// DecodeRLP implements rlp.Decoder
func (tx *RawTransaction) DecodeRLP(s *rlp.Stream) error {
	_, size, _ := s.Kind()
	err := s.Decode(&tx.data)
	if err == nil {
		tx.size.Store(common.StorageSize(rlp.ListSize(size)))
	}
	return err
}

// MarshalJSON encodes the web3 RPC transaction format.
/*func (tx *RawTransaction) MarshalJSON() ([]byte, error) {
	hash := tx.Hash()
	data := tx.data
	data.Hash = &hash
	return data.MarshalJSON()
}*/

// UnmarshalJSON decodes the web3 RPC transaction format.
func (tx *Transaction) UnmarshalJSON(input []byte) error {
	var dec txdata
	if err := dec.UnmarshalJSON(input); err != nil {
		return err
	}
	var V byte
	if isProtectedV(dec.V) {
		chainID := deriveChainId(dec.V).Uint64()
		V = byte(dec.V.Uint64() - 35 - 2*chainID)
	} else {
		V = byte(dec.V.Uint64() - 27)
	}
	if !crypto.ValidateSignatureValues(V, dec.R, dec.S, false) {
		return ErrInvalidSig
	}
	*tx = Transaction{data: dec}
	return nil
}

func (tx *Transaction) Data() []byte       { return common.CopyBytes(tx.data.Payload) }
func (tx *Transaction) Gas() uint64        { return tx.data.GasLimit }
func (tx *Transaction) GasPrice() *big.Int { return new(big.Int).Set(tx.data.Price) }
func (tx *Transaction) Value() *big.Int    { return new(big.Int).Set(tx.data.Amount) }
func (tx *Transaction) Fee() *big.Int {
	if tx.data.Fee == nil {
		return nil
	}
	return new(big.Int).Set(tx.data.Fee)
}
func (tx *Transaction) Nonce() uint64    { return tx.data.AccountNonce }
func (tx *Transaction) CheckNonce() bool { return true }

// To returns the recipient address of the transaction.
// It returns nil if the transaction is a contract creation.
func (tx *Transaction) To() *common.Address {
	if tx.data.Recipient == nil {
		return nil
	}
	to := *tx.data.Recipient
	return &to
}

func (tx *Transaction) Payer() *common.Address {
	if tx.data.Payer == nil {
		return nil
	}
	payer := *tx.data.Payer
	return &payer
}

// Hash hashes the RLP encoding of tx.
// It uniquely identifies the transaction.
func (tx *Transaction) Hash() common.Hash {
	if hash := tx.hash.Load(); hash != nil {
		return hash.(common.Hash)
	}
	v := rlpHash(tx)
	tx.hash.Store(v)
	return v
}

// Size returns the true RLP encoded storage size of the transaction, either by
// encoding and returning it, or returning a previsouly cached value.
func (tx *Transaction) Size() common.StorageSize {
	if size := tx.size.Load(); size != nil {
		return size.(common.StorageSize)
	}
	c := writeCounter(0)
	rlp.Encode(&c, &tx.data)
	tx.size.Store(common.StorageSize(c))
	return common.StorageSize(c)
}

// AsMessage returns the transaction as a core.Message.
//
// AsMessage requires a signer to derive the sender.
//
// XXX Rename message to something less arbitrary?
func (tx *Transaction) AsMessage(s Signer) (Message, error) {
	msg := Message{
		nonce:      tx.data.AccountNonce,
		gasLimit:   tx.data.GasLimit,
		gasPrice:   new(big.Int).Set(tx.data.Price),
		to:         tx.data.Recipient,
		amount:     tx.data.Amount,
		fee:        tx.data.Fee,
		data:       tx.data.Payload,
		checkNonce: true,
	}

	var err error
	msg.from, err = Sender(s, tx)
	if err != nil {
		return msg, err
	}
	return msg, err
}

// WithSignature returns a new transaction with the given signature.
// This signature needs to be formatted as described in the yellow paper (v+27).
func (tx *Transaction) WithSignature(signer Signer, sig []byte) (*Transaction, error) {
	r, s, v, err := signer.SignatureValues(tx, sig)
	if err != nil {
		return nil, err
	}
	cpy := &Transaction{data: tx.data}
	cpy.data.R, cpy.data.S, cpy.data.V = r, s, v
	return cpy, nil
}

func (tx *Transaction) WithSignature_Payment(signer Signer, sig []byte) (*Transaction, error) {
	pr, ps, pv, err := signer.SignatureValues(tx, sig)
	if err != nil {
		return nil, err
	}
	cpy := &Transaction{data: tx.data}
	cpy.data.PR, cpy.data.PS, cpy.data.PV = pr, ps, pv
	return cpy, nil
}

// Cost returns amount + gasprice * gaslimit.
func (tx *Transaction) Cost() *big.Int {
	total := new(big.Int).Mul(tx.data.Price, new(big.Int).SetUint64(tx.data.GasLimit))
	total.Add(total, tx.data.Amount)
	if tx.data.Fee != nil {
		total.Add(total, tx.data.Fee)
	}
	return total
}

// AmountCost returns amount+Fee.
func (tx *Transaction) AmountCost() *big.Int {
	total := big.NewInt(0)
	total.Add(total, tx.data.Amount)
	if tx.data.Fee != nil {
		total.Add(total, tx.data.Fee)
	}
	return total
}

// GasCost returns gasprice * gaslimit.
func (tx *Transaction) GasCost() *big.Int {
	gas := new(big.Int).Mul(tx.data.Price, new(big.Int).SetUint64(tx.data.GasLimit))
	return gas
}

func (tx *Transaction) RawSignatureValues() (*big.Int, *big.Int, *big.Int) {
	return tx.data.V, tx.data.R, tx.data.S
}

func (tx *Transaction) TrueRawSignatureValues() (*big.Int, *big.Int, *big.Int) {
	return tx.data.PV, tx.data.PR, tx.data.PS
}

// Transactions is a Transaction slice type for basic sorting.
type Transactions []*Transaction

// Len returns the length of s.
func (s Transactions) Len() int { return len(s) }

// Swap swaps the i'th and the j'th element in s.
func (s Transactions) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// GetRlp implements Rlpable and returns the i'th element of s in rlp.
func (s Transactions) GetRlp(i int) []byte {
	enc, _ := rlp.EncodeToBytes(s[i])
	return enc
}

// TxDifference returns a new set t which is the difference between a to b.
func TxDifference(a, b Transactions) (keep Transactions) {
	keep = make(Transactions, 0, len(a))

	remove := make(map[common.Hash]struct{})
	for _, tx := range b {
		remove[tx.Hash()] = struct{}{}
	}

	for _, tx := range a {
		if _, ok := remove[tx.Hash()]; !ok {
			keep = append(keep, tx)
		}
	}

	return keep
}

// TxByNonce implements the sort interface to allow sorting a list of transactions
// by their nonces. This is usually only useful for sorting transactions from a
// single account, otherwise a nonce comparison doesn't make much sense.
type TxByNonce Transactions

func (s TxByNonce) Len() int           { return len(s) }
func (s TxByNonce) Less(i, j int) bool { return s[i].data.AccountNonce < s[j].data.AccountNonce }
func (s TxByNonce) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

// TxByPrice implements both the sort and the heap interface, making it useful
// for all at once sorting as well as individually adding and removing elements.
type TxByPrice Transactions

func (s TxByPrice) Len() int           { return len(s) }
func (s TxByPrice) Less(i, j int) bool { return s[i].data.Price.Cmp(s[j].data.Price) > 0 }
func (s TxByPrice) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }

func (s *TxByPrice) Push(x interface{}) {
	*s = append(*s, x.(*Transaction))
}

func (s *TxByPrice) Pop() interface{} {
	old := *s
	n := len(old)
	x := old[n-1]
	*s = old[0 : n-1]
	return x
}

// TransactionsByPriceAndNonce represents a set of transactions that can return
// transactions in a profit-maximizing sorted order, while supporting removing
// entire batches of transactions for non-executable accounts.
type TransactionsByPriceAndNonce struct {
	txs    map[common.Address]Transactions // Per account nonce-sorted list of transactions
	heads  TxByPrice                       // Next transaction for each unique account (price heap)
	signer Signer                          // Signer for the set of transactions
}

// NewTransactionsByPriceAndNonce creates a transaction set that can retrieve
// price sorted transactions in a nonce-honouring way.
//
// Note, the input map is reowned so the caller should not interact any more with
// if after providing it to the constructor.
func NewTransactionsByPriceAndNonce(signer Signer, txs map[common.Address]Transactions) *TransactionsByPriceAndNonce {
	// Initialize a price based heap with the head transactions
	heads := make(TxByPrice, 0, len(txs))
	for from, accTxs := range txs {
		heads = append(heads, accTxs[0])
		// Ensure the sender address is from the signer
		acc, _ := Sender(signer, accTxs[0])
		txs[acc] = accTxs[1:]
		if from != acc {
			delete(txs, from)
		}
	}
	heap.Init(&heads)

	// Assemble and return the transaction set
	return &TransactionsByPriceAndNonce{
		txs:    txs,
		heads:  heads,
		signer: signer,
	}
}

// Peek returns the next transaction by price.
func (t *TransactionsByPriceAndNonce) Peek() *Transaction {
	if len(t.heads) == 0 {
		return nil
	}
	return t.heads[0]
}

// Shift replaces the current best head with the next one from the same account.
func (t *TransactionsByPriceAndNonce) Shift() {
	acc, _ := Sender(t.signer, t.heads[0])
	if txs, ok := t.txs[acc]; ok && len(txs) > 0 {
		t.heads[0], t.txs[acc] = txs[0], txs[1:]
		heap.Fix(&t.heads, 0)
	} else {
		heap.Pop(&t.heads)
	}
}

// Pop removes the best transaction, *not* replacing it with the next one from
// the same account. This should be used when a transaction cannot be executed
// and hence all subsequent ones should be discarded from the same account.
func (t *TransactionsByPriceAndNonce) Pop() {
	heap.Pop(&t.heads)
}

// Message is a fully derived transaction and implements core.Message
//
// NOTE: In a future PR this will be removed.
type Message struct {
	to         *common.Address
	from       common.Address
	payment    common.Address
	nonce      uint64
	amount     *big.Int
	fee        *big.Int
	gasLimit   uint64
	gasPrice   *big.Int
	data       []byte
	checkNonce bool
}

func NewMessage(from common.Address, to *common.Address, payment common.Address, nonce uint64, amount *big.Int, fee *big.Int, gasLimit uint64, gasPrice *big.Int, data []byte, checkNonce bool) Message {
	return Message{
		from:       from,
		to:         to,
		nonce:      nonce,
		amount:     amount,
		gasLimit:   gasLimit,
		gasPrice:   gasPrice,
		data:       data,
		checkNonce: checkNonce,
		payment:    payment,
		fee:        fee,
	}
}

func (m Message) From() common.Address    { return m.from }
func (m Message) To() *common.Address     { return m.to }
func (m Message) Payment() common.Address { return m.payment }
func (m Message) GasPrice() *big.Int      { return m.gasPrice }
func (m Message) Value() *big.Int         { return m.amount }
func (m Message) Fee() *big.Int {
	return m.fee
}
func (m Message) Gas() uint64      { return m.gasLimit }
func (m Message) Nonce() uint64    { return m.nonce }
func (m Message) Data() []byte     { return m.data }
func (m Message) CheckNonce() bool { return m.checkNonce }

type writeCounter common.StorageSize

func (c *writeCounter) Write(b []byte) (int, error) {
	*c += writeCounter(len(b))
	return len(b), nil
}
