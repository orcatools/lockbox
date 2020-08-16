package lockbox

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
)

// Permission defines required levels of access
// type Permission struct {
// }

// A Lockbox provides portable storage for sensitive data.
type Lockbox struct {
	Root       string // path to the lockbox file
	IsOpen     bool
	Namespaces map[string][]*Item
}

// An Item represents anything stored within the Lockbox.
type Item struct {
	// TODO: add validation.
	Path  string `json:"path"`  // path of the given item. Can be path like. eg. "/path/to/some/key.txt"
	Value []byte `json:"value"` // value of the given item.
	// Permissions []*Permission `json:"permissions"` // you can define required permissions for the given item
}

// NewLockbox will return an instance of a lockbox
func NewLockbox(root string) (*Lockbox, error) {
	defaultNamespace := map[string][]*Item{}
	lockbox := &Lockbox{
		Root:       root,
		IsOpen:     true,
		Namespaces: defaultNamespace,
	}
	return lockbox, nil
}

// ListNamespaces will list the defined namespaces within the lockbox.
// Lockbox must be "opened" for this operation.
func (l *Lockbox) ListNamespaces() ([]string, error) {
	return nil, nil
}

// Close will commit the lockbox to disk at its defined root
func (l *Lockbox) Close(key string) error {
	l.IsOpen = false
	data, err := json.Marshal(l)
	if err != nil {
		return err
	}
	edata := encrypt(data, key)
	return ioutil.WriteFile(l.Root, edata, 0644)
}

// OpenLockbox will open the lockbox at the defined root, using the given key.
func OpenLockbox(root, key string) (*Lockbox, error) {
	var l Lockbox
	edata, err := ioutil.ReadFile(root)
	if err != nil {
		return nil, err
	}
	data := decrypt(edata, key)
	err = json.Unmarshal(data, &l)
	if err != nil {
		return nil, err
	}
	l.IsOpen = true
	return &l, nil
}

// AddItem will add an item to the lockbox at the given namespace
func (l *Lockbox) AddItem(namespace string, item *Item) error {
	if l.IsOpen {
		var canAddItem = true
		for _, i := range l.Namespaces[namespace] {
			if i.Path == item.Path {
				canAddItem = false
				break
			}
		}
		if canAddItem {
			l.Namespaces[namespace] = append(l.Namespaces[namespace], item)
			return nil
		}
		return fmt.Errorf("an item at the path %v already exist in namespace %v", item.Path, namespace)
	}
	return errors.New("unable to add item to a lockbox that is closed")
}

// GetItem will get an item from the lockbox
func (l *Lockbox) GetItem(namespace, path string) (*Item, error) {
	if l.IsOpen {
		for _, i := range l.Namespaces[namespace] {
			if i.Path == path {
				return i, nil
			}
		}
		return nil, fmt.Errorf("item not found %v:%v", namespace, path)
	}
	return nil, errors.New("unable to get item from a lockbox that is closed")
}

// DeleteItem will delete an item from the lockbox
func (l *Lockbox) DeleteItem(namespace, path string) error {
	return nil
}

// NewItem will return a new item or error
func NewItem(path string, value []byte) *Item {
	i := &Item{
		Path:  path,
		Value: value,
	}
	// i.Seal("secret42") // TODO: handle this..
	return i
}

// // Seal will encrypt the given item's value
// func (i *Item) Seal(phrase string) {
// 	i.Value = encrypt(i.Value, phrase)
// }

// // Unseal will decrypt the given item's value
// func (i *Item) Unseal(phrase string) {
// 	i.Value = decrypt(i.Value, phrase)
// }
