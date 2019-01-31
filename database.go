package main

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/boltdb/bolt"

	"github.com/sjsafranek/goutils/cryptic"
	// "github.com/sjsafranek/goutils/hashers"
	"github.com/sjsafranek/goutils/minify"

	"github.com/schollz/golock"
)

// // DEFAULT_DB_FILE default database file
const DEFAULT_DB_FILE = "bolt.db"

// // DB_FILE database file to use
var DB_FILE string = DEFAULT_DB_FILE

type Database interface {
	Close()
	Tables() ([]string, error)
	Remove(string, string) error
	CreateTable(string) error
	Get(string, string) (string, error)
	Set(string, string, string) error
	Keys(string) ([]string, error)
}

// Database manages file access through bolt.DB connection and a file lock
type BoltDatabase struct {
	db         *bolt.DB
	glock      *golock.Lock
	passphrase string
}

// Open opens(or creates) bolt database file
func (self *BoltDatabase) Open(db_file string) error {

	logger.Debugf("Opening database: %v", db_file)

	if nil != self.db {
		self.Close()
	}

	if !strings.HasSuffix(db_file, ".db") {
		db_file += ".db"
	}

	// first initiate lockfile
	lock_file := strings.Replace(db_file, ".db", ".lock", -1)
	self.glock = golock.New(
		golock.OptionSetName(lock_file),
		golock.OptionSetInterval(1*time.Millisecond),
		golock.OptionSetTimeout(60*time.Second),
	)

	// lock it
	err := self.glock.Lock()
	if err != nil {
		return err
	}
	//.end

	db, err := bolt.Open(db_file, 0600, &bolt.Options{Timeout: 1 * time.Second})
	self.db = db
	return err
}

// Close close database connection and remove file lock
func (self *BoltDatabase) Close() {
	logger.Warn("Closing database connection")

	self.db.Close()

	// unlock it
	err := self.glock.Unlock()
	if err != nil {
		panic(err)
	}
}

// CreateTable creates a bucket in the bolt database
func (self *BoltDatabase) CreateTable(table_name string) error {
	return self.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(table_name))
		return err
	})
}

// Get retrieves a key from a bucket.
// Decrypts the value using the supplied passphrase.
func (self *BoltDatabase) Get(table, key string) (string, error) {
	if nil == self.db {
		return "", errors.New("Database not opened")
	}
	var result string
	var err error
	return result, self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if nil == b {
			return errors.New("Bucket does not exist")
		}

		v := b.Get([]byte(key))
		decompressed := minify.DecompressByte(v)
		garbage := string(decompressed)
		if "" == garbage {
			return errors.New("Not found")
		}
		result, err = cryptic.Decrypt(self.passphrase, garbage)

		if nil == err && !utf8.ValidString(result) {
			err = errors.New("Not utf-8")
		}

		return err
	})
}

// Set saves a key value to a bucket.
// Encrypts the value using the supplied passphrase.
func (self *BoltDatabase) Set(table, key, value string) error {
	if nil == self.db {
		return errors.New("Database not opened")
	}

	return self.db.Update(func(tx *bolt.Tx) error {
		garbage, err := cryptic.Encrypt(self.passphrase, value)
		if nil != err {
			return err
		}

		b := tx.Bucket([]byte(table))
		if nil == b {
			return errors.New("Bucket does not exist")
		}

		compressed := minify.CompressByte([]byte(garbage))
		return b.Put([]byte(key), compressed)
	})
}

// Keys lists all keys with in a bucket
func (self *BoltDatabase) Keys(table string) ([]string, error) {
	var result []string
	if nil == self.db {
		return result, errors.New("Database not opened")
	}
	return result, self.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(table))
		if nil == b {
			return errors.New("Bucket does not exist")
		}
		return b.ForEach(func(k, v []byte) error {
			result = append(result, string(k))
			return nil
		})
	})
}

// Remove deletes a key from a bucket
func (self *BoltDatabase) Remove(table string, key string) error {
	return self.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(table))
		if bucket == nil {
			return fmt.Errorf("Bucket %q not found!", table)
		}

		err := bucket.Delete([]byte(key))
		if err != nil {
			return fmt.Errorf("Could not delete key: %s", err)
		}
		return err
	})
}

// Tables returns list of buckets
func (self *BoltDatabase) Tables() ([]string, error) {
	var result []string
	if nil == self.db {
		return result, errors.New("Database not opened")
	}
	return result, self.db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			result = append(result, string(name))
			return nil
		})
	})
}

// OpenDb opens bolt file and returns Database
func NewDatabase(db_file string, passphrase string) (Database, error) {
	// passphrase = hashers.Sha512HashString(passphrase)
	db := BoltDatabase{passphrase: passphrase}
	err := db.Open(db_file)
	if nil != err {
		return &db, err
	}
	return &db, err
}
