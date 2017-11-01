package main

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"

	"github.com/Rudd-O/curvetls"
)

var _ = errors.Wrap

type Keypair struct {
	Pub  curvetls.Pubkey
	Priv curvetls.Privkey
}

func Manage(ip string) error {

	// generate a keypair
	priv, pub, err := curvetls.GenKeyPair()
	if err != nil {
		return errors.Wrap(err, "could not generate key pair")
	}

	pair := Keypair{pub, priv}

	keystorePath := filepath.Join(os.Getenv("HOME"), ".config", "hpt", "keys.db")

	// store in boltdb
	db, err := bolt.Open(keystorePath, os.FileMode(os.O_RDWR), nil)
	if err != nil {
		return errors.Wrap(err, "could not open keys.db")
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucketIfNotExists([]byte("keys"))
		if err != nil {
			return err
		}
		encoded, err := gobEncode(&pair)
		if err != nil {
			return err
		}
		// keyed by ip address; long term we'll want something different
		if err := b.Put([]byte(ip), encoded); err != nil {
			return err
		}
		return nil
	})

	fmt.Println("added keypair for", ip)

	return nil
}

func gobEncode(v interface{}) ([]byte, error) {
	// Note: v must be a pointer
	var buf []byte
	b := bytes.NewBuffer(buf)
	enc := gob.NewEncoder(b)

	err := enc.Encode(v)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

func gobDecode(data []byte, v interface{}) error {
	// Note: v must be a pointer
	b := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(b)
	return decoder.Decode(v)

}
