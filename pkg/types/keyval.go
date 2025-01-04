package types

import "fmt"

type KeyVal struct {
	Key   string
	Value string
}

func (k *KeyVal) String() string {
	return fmt.Sprintf("%s=%s", k.Key, k.Value)
}
