package lockbox

import (
	"errors"
	"fmt"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	bolt "go.etcd.io/bbolt"
)

const (
	lockboxbucket = "lockbox"
)

// A Lockbox provides portable, secure, storage for secrets.
type Lockbox struct {
	Name      string
	MasterKey string // this is the key used for initial encryption/decryption
	Store     *bolt.DB
	Locked    bool
	OTPKey    *otp.Key // this is the totp key
}

// GetLockbox will return an instance of lockbox
func GetLockbox(name, master string) (*Lockbox, error) {
	db, err := bolt.Open(fmt.Sprintf("%v.lockbox", name), 0600, nil)
	if err != nil {
		return nil, err
	}

	// create our default bucket
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(lockboxbucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	lockbox := &Lockbox{
		Name:      name,
		Store:     db,
		Locked:    true,
		MasterKey: master,
	}
	// NOTE: should encrypt it here?
	return lockbox, nil
}

// Close will close the lockbox.
func (l *Lockbox) Close() error {
	// TODO: lock, serialize, etc, before close.

	return l.Store.Close()
}

// Init will perform initialization operations on the given lockbox
func (l *Lockbox) Init(namespace string) error {
	// write some data to the /lockbox/meta path
	// eg: date created, date modifed, lockbox version
	// eg: user defined access policies
	// eg: totp meta data
	// eg: other things we think of ?
	created, err := time.Now().MarshalBinary()
	if err != nil {
		return err
	}

	// updates the lockbox metadata for a given namespace
	err = l.Store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(lockboxbucket))
		err := b.Put([]byte(fmt.Sprintf("/lockbox/meta/%v/date/created", namespace)), created)
		return err
	})
	if err != nil {
		return err
	}

	// generate the otp key
	key, err := generateOTPKey(namespace)
	fmt.Println("OTP SECRET:", key.Secret())
	if err != nil {
		return err
	}

	// encrypt the otp key with the master key
	enckey, err := encrypt([]byte(key.String()), l.MasterKey)
	if err != nil {
		return err
	}
	// fmt.Println("ENCKEY SIZE:", len(enckey))

	// store encrypted otp key within metadata for the namespace
	err = l.Store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(lockboxbucket))
		err := b.Put([]byte(fmt.Sprintf("/lockbox/meta/%v/otp/key", namespace)), enckey)
		return err
	})

	return nil
}

// GetMetaData will return lockbox metadata for given path
func (l *Lockbox) GetMetaData(path string) ([]byte, error) {
	var meta []byte
	metapath := fmt.Sprintf("/lockbox/meta/%s", path)
	l.Store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(lockboxbucket))
		meta = b.Get([]byte(metapath))
		return nil
	})
	if len(meta) > 0 {
		return meta, nil
	}
	return nil, errors.New("invalid metadata path")
}

// Lock will lock the given lockbox
func (l *Lockbox) Lock() error {
	l.Locked = true
	return nil
}

// Unlock will unlock the given lockbox, as long as the provided key[s] are valid.
func (l *Lockbox) Unlock(namespace, code string) error {
	enckey, err := l.GetMetaData(fmt.Sprintf("%v/otp/key", namespace))
	// fmt.Println("ENCKEY SIZE:", len(enckey))
	if err != nil {
		return err
	}
	keyurl, err := decrypt(enckey, l.MasterKey)
	if err != nil {
		return err
	}
	key, err := otp.NewKeyFromURL(string(keyurl))
	if err != nil {
		return err
	}

	// fmt.Println(key.AccountName(), key.Secret())

	valid := totp.Validate(code, key.Secret())

	// // if the key is valid, then locked should be false.
	l.Locked = !valid
	return nil
}

// SetValue will set a value at a given path
func (l *Lockbox) SetValue(value []byte, namespace, path string) error {
	if l.Locked {
		return errors.New("cannot set value while lockbox is locked")
	}

	// encrypt the value with the master key
	// TODO: this only offers 1 layer of security.
	// we need to figure out a smarter way to protect the values.
	encval, err := encrypt(value, l.MasterKey)
	if err != nil {
		return err
	}

	return l.Store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(lockboxbucket))
		err := b.Put([]byte(fmt.Sprintf("/lockbox/value/%v/%v", namespace, path)), encval)
		return err
	})
}

// GetValue will get a value from a given path
func (l *Lockbox) GetValue(namespace, path string) ([]byte, error) {
	if l.Locked {
		return nil, errors.New("cannot get value from a lockbox that is locked")
	}
	var encvalue []byte
	datapath := fmt.Sprintf("/lockbox/value/%s/%s", namespace, path)
	l.Store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(lockboxbucket))
		encvalue = b.Get([]byte(datapath))
		return nil
	})
	if len(encvalue) > 0 {
		value, err := decrypt(encvalue, l.MasterKey)
		if err != nil {
			return nil, err
		}
		return value, nil
	}
	return nil, errors.New("invalid data path")
}

// RemValue will remove the value at a given path
func (l *Lockbox) RemValue(path string) error {
	return nil
}
