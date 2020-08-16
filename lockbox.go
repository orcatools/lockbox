package lockbox

import (
	"errors"
	"io/ioutil"
)

// A Lockbox provides portable storage for sensitive data.
type Lockbox struct {
	Root   string  // path to the lockbox file
	kv     *KVFile // underlying storage for Lockbox
	Status int     // 0 if open, 1 if closed
}

// An Item represents anything stored within the Lockbox.
type Item struct {
	Namespace string
	Path      string
	Value     string
}

// NewLockbox will return an instance of a lockbox
func NewLockbox(root string) (*Lockbox, error) {
	kv, err := NewKVFile(root)
	if err != nil {
		return nil, err
	}
	lockbox := &Lockbox{
		Root:   root,
		kv:     kv,
		Status: 0,
	}
	return lockbox, nil
}

// OpenLockbox will attempt to open a lockbox at the specified root
func OpenLockbox(root, password string) (*Lockbox, error) {
	data := decryptFile(root, password)
	err := ioutil.WriteFile(root, data, 0644)
	if err != nil {
		return nil, err
	}
	kv, err := NewKVFile(root)
	if err != nil {
		return nil, err
	}
	lockbox := &Lockbox{
		Root:   root,
		kv:     kv,
		Status: 0,
	}
	return lockbox, nil
}

// Lock will lock the lockbox
func (l *Lockbox) Lock(password string) error {
	l.Status = 1
	err := l.kv.Close()
	if err != nil {
		return err
	}
	data, err := ioutil.ReadFile(l.Root)
	if err != nil {
		return err
	}
	encryptFile(l.Root, data, password)
	return err
}

// Unlock will unlock the lockbox
// func (l *Lockbox) Unlock(password string) error {
// 	l.Status = 0

// 	return err
// }

// AddItem will add an item to the lockbox
func (l *Lockbox) AddItem(namespace, path, value string) error {
	if l.Status != 1 {
		l.kv.Put(namespace, path, value)
		return nil
	}
	return errors.New("unable to add item to a lockbox that is locked")
}

// GetItem will get an item from the lockbox
func (l *Lockbox) GetItem(namespace, path string) (*Item, error) {
	if l.Status != 1 {
		val, err := l.kv.Get(namespace, path)
		if err != nil {
			return nil, err
		}
		i := &Item{
			Namespace: namespace,
			Path:      path,
			Value:     val,
		}
		return i, nil
	}
	return nil, errors.New("unable to get item from a lockbox that is locked")
}

// DeleteItem will delete an item from the lockbox
func (l *Lockbox) DeleteItem(namespace, path string) error {
	return nil
}
