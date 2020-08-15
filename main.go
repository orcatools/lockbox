package main

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

func main() {

	// pass := []byte("helloworld")
	// armor, err := helper.EncryptMessageWithPassword(pass, "some secret text")
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(armor)
	// message, err := helper.DecryptMessageWithPassword(pass, armor)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(message)
	kv, err := New("lockbox")
	if err != nil {
		logrus.Errorf(err.Error())
	}
	// kv.Put("dan", "keys/meh/1", "World")
	// kv.Put("dan", "keys/meh/2", "Foo")
	// kv.Put("dan", "keys/foo/1", "Bar")
	// val, err := kv.Get("dan", "keys/meh/2")
	// fmt.Println(val)
	m, err := kv.GetAll("dan")
	for k, v := range m {
		fmt.Println(k, v)
	}

	err = kv.Close()
	if err != nil {
		logrus.Errorf(err.Error())
	}
}
