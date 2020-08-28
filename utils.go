package lockbox

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

func createHash(key string) string {
	hasher := md5.New()
	hasher.Write([]byte(key))
	return hex.EncodeToString(hasher.Sum(nil))
}

func encrypt(data []byte, passphrase string) ([]byte, error) {
	block, _ := aes.NewCipher([]byte(createHash(passphrase)))
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ciphertext := gcm.Seal(nonce, nonce, data, nil)
	return ciphertext, nil
}

func decrypt(data []byte, passphrase string) ([]byte, error) {
	key := []byte(createHash(passphrase))
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}

func encryptFile(filename string, data []byte, passphrase string) error {
	f, _ := os.Create(filename)
	defer f.Close()
	enc, err := encrypt(data, passphrase)
	if err != nil {
		return err
	}
	f.Write(enc)
	return nil
}

func decryptFile(filename string, passphrase string) ([]byte, error) {
	data, _ := ioutil.ReadFile(filename)
	return decrypt(data, passphrase)
}

// generate a OTP key for a given namespace
func generateOTPKey(namespace string) (*otp.Key, error) {
	return totp.Generate(totp.GenerateOpts{
		Issuer:      "lockbox",
		AccountName: fmt.Sprintf("%v@lockbox", namespace),
	})
}
