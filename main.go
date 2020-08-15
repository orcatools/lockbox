package main

import "fmt"

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
	kv, err := Open("lockbox")
	if err != nil {
		panic(err)
	}
	kv.Put("dan", "Hello", "World")
	fmt.Println(kv.CountKeys("dan"))
	err = kv.Close()
	if err != nil {
		panic(err)
	}
}
