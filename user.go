package lockbox

import (
	"crypto/sha512"

	"golang.org/x/crypto/pbkdf2"
)

// A User is someone who would access the Lockbox
type User struct {
	Username      string
	Password      string
	EncryptionKey string
}

// GetUserEncryptionKey will use pbkdf2 to return a
func (u *User) GetUserEncryptionKey() []byte {
	return pbkdf2.Key([]byte(u.Password), []byte(u.Username), 1024, 128, sha512.New)
}

// TODO: add a validate method?
