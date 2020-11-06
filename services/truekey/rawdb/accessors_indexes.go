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
	"ethereum/keyservice/common"
)

// WriteAccountLookupEntries stores a positional metadata for every account from
// a wallet, enabling hash based account  lookups.

// DeleteAccountLookupEntry removes all account data associated with a hash.
func DeleteAccountLookupEntry(db DatabaseDeleter, hash common.Hash) {
	db.Delete(accountLookupKey(hash))
}

// HasAccountLookupEntry verifies the existence of a accountLook entry corresponding to the hash.
func HasAccountLookupEntry(db DatabaseReader, hash common.Hash) bool {
	if has, err := db.Has(accountLookupKey(hash)); !has || err != nil {
		return false
	}
	return true
}
