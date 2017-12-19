package config

import (
	"fmt"

	"github.com/go-ini/ini"
)

func GetaKey(f *ini.File, section string, key string) string {
	val, err := f.Section(section).GetKey(key)
	if err != nil {
		fmt.Println(err)
	}
	return val.String()
}
