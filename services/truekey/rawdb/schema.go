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

// Package rawdb contains a collection of low level database accessors.
package rawdb

import (
	"ethereum/keyservice/common"
)

// The fields below define the low level database schema prefixing.
var (
	// fastTrieProgressKey tracks the number of trie entries imported during fast sync.
	indexKey = []byte("index")

	accountLookupPrefix = []byte("l") // accountLookupPrefix + hash -> account lookup metadata

	// Data item prefixes (use single byte to avoid mixing data types, avoid `i`, used for indexes).
	adminPrefix       = []byte("a") // headerPrefix + num (uint64 big endian) + hash -> header
	adminWalletPrefix = []byte("b") // adminWalletPrefix  + hash -> adminWallet
	dappInfoPrefix    = []byte("c") // dappInfoPrefix  + hash -> dappInfo
	adminInfoPrefix   = []byte("d") // adminInfoPrefix + hash -> header
	accountPrefix     = []byte("e") // dappPrefix + hash (dappid) + root -> dapp account index
)

// AccountLookup is a positional metadata to help looking up the data content of
// a account given only its hash.
type AccountLookup struct {
	WalletHash common.Hash
	Index      uint64
}

// headerKey = headerPrefix + hash
func IndexKey(hash common.Hash) []byte {
	return append(indexKey, hash.Bytes()...)
}

// accountIndexKey = accountPrefix + root + dappid
func accountIndexKey(root, dappid common.Hash) []byte {
	return append(append(accountPrefix, root.Bytes()...), dappid.Bytes()...)
}

// headerKey = headerPrefix + hash
func adminKey(root, dappid common.Hash) []byte {
	return append(append(adminPrefix, root.Bytes()...), dappid.Bytes()...)
}

// adminWalletKey = adminWalletPrefix + hash
func adminWalletKey(hash common.Hash) []byte {
	return append(adminWalletPrefix, hash.Bytes()...)
}

// adminWalletKey = adminWalletPrefix + hash
func dappInfoKey(hash common.Hash) []byte {
	return append(dappInfoPrefix, hash.Bytes()...)
}

// accountLookupKey = accountLookupPrefix + hash
func accountLookupKey(hash common.Hash) []byte {
	return append(accountLookupPrefix, hash.Bytes()...)
}

// adminWalletKey = adminWalletPrefix + hash
func adminInfoKey(hash common.Hash) []byte {
	return append(adminInfoPrefix, hash.Bytes()...)
}
