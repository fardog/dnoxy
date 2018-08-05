package cmd

import (
	"fmt"
	"strings"
)

type KeyValue map[string][]string

func (k KeyValue) Set(kv string) error {
	kvs := strings.SplitN(kv, "=", 2)
	if len(kvs) != 2 {
		return fmt.Errorf("invalid format for %v; expected KEY=VALUE", kv)
	}
	key, value := kvs[0], kvs[1]

	if vs, ok := k[key]; ok {
		k[key] = append(vs, value)
	} else {
		k[key] = []string{value}
	}

	return nil
}

func (k KeyValue) String() string {
	var s []string
	for k, vs := range k {
		for _, v := range vs {
			s = append(s, fmt.Sprintf("%v=%v", k, v))
		}
	}

	return strings.Join(s, " ")
}

type Values []string

func (v *Values) Set(n string) error {
	*v = append(*v, n)
	return nil
}

func (v *Values) String() string {
	return strings.Join([]string(*v), ",")
}
