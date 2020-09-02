package lockbox

import (
	"errors"
	"fmt"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
	bolt "go.etcd.io/bbolt"
)

const (
	lockboxbucket = "lockbox"
)

// A Lockbox provides portable, secure, storage for secrets.
type Lockbox struct {
	Name             string
	Store            *bolt.DB
	Locked           bool
	CurrentUser      *User
	CurrentNamespace string
}

// GetLockbox will return an instance of lockbox
func GetLockbox(lbname string) (*Lockbox, error) {
	db, err := bolt.Open(lbname, 0600, nil)
	if err != nil {
		return nil, err
	}

	// create our default buckets
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
		Name:   lbname,
		Store:  db,
		Locked: true,
	}

	return lockbox, nil
}

// Close will close the lockbox.
func (l *Lockbox) Close() error {
	l.Locked = true
	return l.Store.Close()
}

// Init will perform initialization operations on the given lockbox
func (l *Lockbox) Init(namespace, username, password string) (*otp.Key, error) {

	// make sure the bucket exist to contain our namespace
	err := l.Store.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(namespace))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// we need a user to initialize the lockbox
	// NOTE: should this user default to "root"?
	user := &User{
		Username: username,
		Password: password,
	}
	// TODO: validate user struct

	err = l.Store.Update(func(tx *bolt.Tx) error {
		// we write metadata to the lockbox bucket, which is stored seperately from
		// the acutal data. This should increase the security of lockbox by seperating
		// this data, from secret data.
		b := tx.Bucket([]byte(lockboxbucket))
		usernameHash := createHash(username)
		encPassword, err := encrypt([]byte(password), usernameHash)
		if err != nil {
			return err
		}
		err = b.Put([]byte(fmt.Sprintf("/lockbox/meta/%v/users/%v", namespace, usernameHash)), []byte(encPassword))
		return err
	})

	// generate the otp key
	otpkey, err := generateOTPKey(namespace)
	if err != nil {
		return nil, err
	}

	// encrypt the otp key with the user's encryption key
	encotpkey, err := encrypt([]byte(otpkey.String()), string(user.GetUserEncryptionKey()))
	if err != nil {
		return nil, err
	}

	// store encrypted otp key within metadata for the namespace
	err = l.Store.Update(func(tx *bolt.Tx) error {
		// we write metadata to the lockbox bucket, which is stored seperately from
		// the acutal data. This should increase the security of lockbox by seperating
		// this data, from secret data.
		b := tx.Bucket([]byte(lockboxbucket))
		err := b.Put([]byte(fmt.Sprintf("/lockbox/meta/%v/otp/key", namespace)), encotpkey)
		return err
	})
	if err != nil {
		return nil, err
	}

	// NOTE: it is the caller's job to retrieve the secret from the otpkey.
	return otpkey, nil
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
func (l *Lockbox) Unlock(namespace, username, password, code string) error {
	// get our encrypted otp key from the metadata store
	encotpkey, err := l.GetMetaData(fmt.Sprintf("%v/otp/key", namespace))
	if err != nil {
		return err
	}

	user := &User{
		Username: username,
		Password: password,
	}

	// pw is the stored pw of the given username, we need to validate it against the provided password.
	usernameHash := createHash(user.Username)
	encpw, err := l.GetMetaData(fmt.Sprintf("%v/users/%v", namespace, usernameHash))
	if err != nil {
		return errors.New("invalid username")
	}

	// user provided invalid password for given username
	pw, err := decrypt(encpw, usernameHash)
	if err != nil {
		return errors.New("unable to unseal user metadata")
	}

	if password != string(pw) {
		return errors.New("invalid password")
	}

	user.EncryptionKey = string(user.GetUserEncryptionKey())

	keyurl, err := decrypt(encotpkey, user.EncryptionKey)
	if err != nil {
		return err
	}
	otpkey, err := otp.NewKeyFromURL(string(keyurl))
	if err != nil {
		return err
	}
	valid := totp.Validate(code, otpkey.Secret())

	// if the key is valid, then locked should be false.
	l.Locked = !valid
	l.CurrentUser = user
	l.CurrentNamespace = namespace
	return nil
}

// SetValue will set a value at a given path
func (l *Lockbox) SetValue(path, value []byte) error {
	if l.Locked {
		return errors.New("cannot set value while lockbox is locked")
	}

	// encrypt the provided value with the current users encryption key.
	encval, err := encrypt(value, l.CurrentUser.EncryptionKey)
	if err != nil {
		return err
	}

	// store the encrypted value at the desired path
	return l.Store.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(l.CurrentNamespace))
		err := b.Put(path, encval)
		return err
	})
}

// GetValue will get a value from a given path
func (l *Lockbox) GetValue(path []byte) ([]byte, error) {
	if l.Locked {
		return nil, errors.New("cannot get value from a lockbox that is locked")
	}

	var encvalue []byte
	l.Store.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(l.CurrentNamespace))
		encvalue = b.Get(path)
		return nil
	})
	if len(encvalue) > 0 {
		value, err := decrypt(encvalue, l.CurrentUser.EncryptionKey)
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
