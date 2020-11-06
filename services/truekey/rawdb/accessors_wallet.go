// Copyright 2018 The go-ethereum Authors
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

package rawdb

import (
	"bytes"
	"ethereum/keyservice/common"
	"ethereum/keyservice/log"
	"ethereum/keyservice/rlp"
	"ethereum/keyservice/services/truekey/types"
	"math/big"
)

// ReadIndexKey retrieves the number of account path index
func ReadIndexKey(db DatabaseReader, key common.Hash) uint64 {
	data, _ := db.Get(IndexKey(key))
	if len(data) == 0 {
		return 0
	}
	return new(big.Int).SetBytes(data).Uint64()
}

// WriteIndexKey stores the number of account path index
func WriteIndexKey(db DatabaseWriter, key common.Hash, count uint64) {
	if err := db.Put(IndexKey(key), new(big.Int).SetUint64(count).Bytes()); err != nil {
		log.Crit("Failed to store dapp index", "err", err)
	}
}

// ReadCanonicalHash retrieves the hash assigned to a canonical block number.
func ReadAdminPassword(db DatabaseReader, root, dappid common.Hash) []common.Address {
	data, _ := db.Get(adminKey(root, dappid))
	if len(data) == 0 {
		return []common.Address{}
	}
	var admins []common.Address
	if err := rlp.Decode(bytes.NewReader(data), &admins); err != nil {
		log.Error("Invalid admin password RLP", "hash", root, "err", err)
		return nil
	}

	return admins
}

// WriteAdminPassword stores the hash assigned to a admin password.
func WriteAdminPassword(db DatabaseWriter, root, dappid common.Hash, admins []common.Address) {
	data, err := rlp.EncodeToBytes(admins)
	if err != nil {
		log.Crit("Failed to RLP encode admins", "err", err)
	}
	if err := db.Put(adminKey(root, dappid), data); err != nil {
		log.Crit("Failed to store admin hash", "err", err)
	}
}

// HasAdminPassword verifies the existence of a admin password corresponding to the hash.
func HasAdminPassword(db DatabaseReader, root, dappid common.Hash) bool {
	if has, err := db.Has(adminKey(root, dappid)); !has || err != nil {
		return false
	}
	return true
}

// ReadAdminWalletRLP retrieves the admin wallet in RLP encoding.
func readAdminWalletRLP(db DatabaseReader, hash common.Hash) rlp.RawValue {
	data, _ := db.Get(adminWalletKey(hash))
	return data
}

// WriteAdminWalletRLP stores an RLP encoded admin wallet into the database.
func writeAdminWalletRLP(db DatabaseWriter, hash common.Hash, rlp rlp.RawValue) {
	if err := db.Put(adminWalletKey(hash), rlp); err != nil {
		log.Crit("Failed to store admin wallet", "err", err)
	}
}

// HasAdminWallet verifies the existence of a admin wallet corresponding to the hash.
func HasAdminWallet(db DatabaseReader, hash common.Hash) bool {
	if has, err := db.Has(adminWalletKey(hash)); !has || err != nil {
		return false
	}
	return true
}

// ReadAdminWallet retrieves the admin wallet corresponding to the hash.
func ReadAdminWallet(db DatabaseReader, hash common.Hash) *types.AdminWallet {
	data := readAdminWalletRLP(db, hash)
	if len(data) == 0 {
		return nil
	}
	body := new(types.AdminWallet)
	if err := rlp.Decode(bytes.NewReader(data), body); err != nil {
		log.Error("Invalid block body RLP", "hash", hash, "err", err)
		return nil
	}
	return body
}

// WriteAdminWallet store a admin wallet into the database.
func WriteAdminWallet(db DatabaseWriter, hash common.Hash, wallet *types.AdminWallet) {
	data, err := rlp.EncodeToBytes(wallet)
	if err != nil {
		log.Crit("Failed to RLP encode admin wallet", "err", err)
	}
	writeAdminWalletRLP(db, hash, data)
}

// DeleteAdminWallet removes admin wallet data associated with a hash.
func DeleteAdminWallet(db DatabaseDeleter, hash common.Hash) {
	if err := db.Delete(adminWalletKey(hash)); err != nil {
		log.Crit("Failed to delete admin wallet", "err", err)
	}
}

// ReadAdminWalletRLP retrieves the admin wallet in RLP encoding.
func readDappInfoRLP(db DatabaseReader, hash common.Hash) rlp.RawValue {
	data, _ := db.Get(dappInfoKey(hash))
	return data
}

// WriteAdminWalletRLP stores an RLP encoded dapp info into the database.
func writeDappInfoRLP(db DatabaseWriter, hash common.Hash, rlp rlp.RawValue) {
	if err := db.Put(dappInfoKey(hash), rlp); err != nil {
		log.Crit("Failed to store dapp info", "err", err)
	}
}

// HasChildAccount verifies the existence of a dapp info corresponding to the hash.
func HasChildAccount(db DatabaseReader, hash common.Hash) bool {
	if has, err := db.Has(dappInfoKey(hash)); !has || err != nil {
		return false
	}
	return true
}

// ReadChildAccount retrieves the admin wallet corresponding to the hash.
func ReadChildAccount(db DatabaseReader, hash common.Hash) *types.ChildAccount {
	data := readDappInfoRLP(db, hash)
	if len(data) == 0 {
		return nil
	}
	body := new(types.ChildAccount)
	if err := rlp.Decode(bytes.NewReader(data), body); err != nil {
		log.Error("Invalid dapp info RLP", "hash", hash, "err", err)
		return nil
	}
	return body
}

// WriteChildAccount store a dapp info into the database.
func WriteChildAccount(db DatabaseWriter, hash common.Hash, wallet *types.ChildAccount) {
	data, err := rlp.EncodeToBytes(wallet)
	if err != nil {
		log.Crit("Failed to RLP encode dapp info", "err", err)
	}
	writeDappInfoRLP(db, hash, data)
}

// DeleteDappInfo removes dapp info data associated with a hash.
func DeleteChildAccount(db DatabaseDeleter, hash common.Hash) {
	if err := db.Delete(dappInfoKey(hash)); err != nil {
		log.Crit("Failed to delete dapp info", "err", err)
	}
}

// ReadAdminInfo retrieves the hash assigned to a canonical block number.
func ReadRootInfo(db DatabaseReader, key common.Hash) []common.Hash {
	data, _ := db.Get(adminInfoKey(key))
	if len(data) == 0 {
		return []common.Hash{}
	}
	var admins []common.Hash
	if err := rlp.Decode(bytes.NewReader(data), &admins); err != nil {
		log.Error("Invalid root info RLP", "hash", key, "err", err)
		return nil
	}

	return admins
}

// WriteAdminInfo stores the hash assigned to a admin info.
func WriteRootInfo(db DatabaseWriter, key common.Hash, admins []common.Hash) {
	data, err := rlp.EncodeToBytes(admins)
	if err != nil {
		log.Crit("Failed to RLP encode root info", "err", err)
	}
	if err := db.Put(adminInfoKey(key), data); err != nil {
		log.Crit("Failed to store root info", "err", err)
	}
}

// HasAdminInfo verifies the existence of a admin info corresponding to the hash.
func HasRootInfo(db DatabaseReader, hash common.Hash) bool {
	if has, err := db.Has(adminInfoKey(hash)); !has || err != nil {
		return false
	}
	return true
}
