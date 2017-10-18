package models

type VaultKeyPair struct {
	PubKey []byte            `json:"pub_key"`
	Keys   map[string][]byte `json:"keys"`
}
