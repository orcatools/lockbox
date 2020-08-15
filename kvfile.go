package main

import (
	"encoding/gob"
	"os"
)

// KVFile offers minimal "kv store" features, but in a single file.
// The focus here was more on functionality, and not performance.
// It was designed to fit a specific need, and only focuses on that need.
type KVFile struct {
	Path    string // path to the kvfile
	Entries map[string]map[string]string
}

// New will create a new KVFile instance
func New(p string) (*KVFile, error) {
	k := &KVFile{
		Path:    p,
		Entries: make(map[string]map[string]string),
	}
	if _, err := os.Stat(p); err == nil || os.IsExist(err) {
		f, err := os.OpenFile(p, os.O_RDONLY, 0644)
		fs, err := f.Stat()
		if fs.Size() > 0 {
			dec := gob.NewDecoder(f)
			err = dec.Decode(&k.Entries)
			if err != nil {
				return nil, err
			}
		}
	}
	return k, nil
}

// Close will commit the changes to disk.
func (kv *KVFile) Close() error {
	f, err := os.OpenFile(kv.Path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	enc := gob.NewEncoder(f)
	err = enc.Encode(kv.Entries)
	return err
}

// Put will put the key/value pair into the KVFile
// This method appends the given entry into the kvfile
func (kv *KVFile) Put(bucket, key, value string) {
	// TODO: check if key is there? or just always overwrite?

	if kv.Entries[bucket] == nil {
		kv.Entries[bucket] = make(map[string]string)
	}
	kv.Entries[bucket][key] = value

	// TODO: add any error checking to see if value is in map after PUT?
}

// Get will return the value of a given key from within the kvfile
func (kv *KVFile) Get(bucket, key string) (val string, err error) {
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
