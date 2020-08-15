package main

import (
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
	kv.Put("dan", "Hello", "World")
	// val, err := kv.Get("dan", "Hello")
	// fmt.Println(val)
	err = kv.Close()
	if err != nil {
		logrus.Errorf(err.Error())
	}
}
