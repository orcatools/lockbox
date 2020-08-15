package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
)

// KVFile offers minimal "kv store" features, but in a single file.
// The focus here was more on functionality, and not performance.
// It was designed to fit a specific need, and only focuses on that need.
type KVFile struct {
	Path    string // path to the kvfile
	File    *os.File
	Entries map[string]map[string]string `json:"entries"`
}

// Open will open the file for writing
func Open(p string) (*KVFile, error) {
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY, 0644)
	var entries map[string]map[string]string

	if err != nil {
		return nil, err
	}
	k := &KVFile{
		Path: p,
		File: f,
	}
	k.Entries = make(map[string]map[string]string)
	data, err := ioutil.ReadFile(p)
	if err != nil {
		return nil, err
	}
	if len(data) > 0 {
		err = json.Unmarshal(data, &entries)
		if err != nil {
			return nil, err
		}
		k.Entries = entries
	}
	return k, nil
}

// Close will commit the changes to disk.
func (kv *KVFile) Close() error {
	defer kv.File.Close()
	data, err := json.Marshal(kv.Entries)
	if err != nil {
		return err
	}
	n, err := kv.File.Write(data)
	if err != nil {
		return err
	}
	if n != len(data) {
		return errors.New("KVFile ERROR: expected byte size written to disk did not match")
	}
	return nil
}

// Put will put the key/value pair into the KVFile
// This method appends the given entry into the kvfile
func (kv *KVFile) Put(bucket, key string, value string) error {
	// TODO: check if key is there
	// TODO: initialize map
	if kv.Entries[bucket] == nil {
		kv.Entries[bucket] = make(map[string]string)
	}
	kv.Entries[bucket][key] = value
	return nil
}

// Get will return the value of a given key from within the kvfile
func (kv *KVFile) Get(bucket, key string) (val interface{}, err error) {
	val = kv.Entries[bucket][key]
	return val, nil
}

// Delete will delete the value of a given key from within the kvfile
func (kv *KVFile) Delete(bucket, key string) {
	delete(kv.Entries[bucket], key)
	if len(kv.Entries[bucket]) == 0 {
		delete(kv.Entries, bucket)
	}
	// TODO: consider adding some error handling to verify the entries were removed?
}

// CountKeys will return the number of keys stored within the given bucket
func (kv *KVFile) CountKeys(bucket string) (count int) {
	count = len(kv.Entries[bucket])
	return
}
