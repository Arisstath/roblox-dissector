package datamodel

import (
	"encoding/hex"
	"math/rand"
)

// TODO: should work with PeerID
type InstanceDictionary struct {
	Scope         string
	InstanceIndex uint32
}

func NewInstanceDictionary() *InstanceDictionary {
	scope := make([]byte, 0x10)
	n, err := rand.Read(scope)
	if n < 0x10 && err != nil {
		panic(err)
	}
	scopeStr := "RBX" + hex.EncodeToString(scope)

	// We must start at instanceindex 1, because 0 is reserved for NULL
	return &InstanceDictionary{Scope: scopeStr, InstanceIndex: 1}
}

func (dictionary *InstanceDictionary) NewReference() Reference {
	reference := Reference{Scope: dictionary.Scope, Id: dictionary.InstanceIndex}
	dictionary.InstanceIndex++
	return reference
}
