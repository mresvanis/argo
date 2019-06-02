package registry

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	bolt "go.etcd.io/bbolt"
	"golang.org/x/xerrors"
)

var (
	ErrOpen   = xerrors.New("could not open registry")
	ErrUpdate = xerrors.New("could not update registry")
	ErrGet    = xerrors.New("could not get from registry")
)

// Registrar keeps the file input and offsets registry.
type Registrar interface {
	// Open the bolt db file or return an error on failure.
	Open() error

	// Close the bolt db file.
	Close()

	// GetOffset returns the stored offset for the specified file path.
	GetOffset(string) (int64, error)

	// UpdateOffset updates the offset for the specified key or returns
	// an error on failure.
	UpdateOffset(key string, offset int64) error
}

type Registry struct {
	db   *bolt.DB
	log  *log.Logger
	path string
	term chan struct{}

	bucketName []byte
}

func NewRegistry(path string) Registrar {
	reg := new(Registry)

	reg.path = path
	reg.bucketName = []byte("argo")
	reg.log = log.New(os.Stderr, fmt.Sprintf("[reg] %s ", reg.path), log.LstdFlags)

	return reg
}

func (reg *Registry) Open() error {
	db, err := bolt.Open(reg.path, 0600, &bolt.Options{Timeout: 2 * time.Second})
	if err != nil {
		return xerrors.Errorf("%w: %s", ErrOpen, err.Error())
	}

	reg.db = db

	return nil
}

func (reg *Registry) Close() {
	reg.db.Close()
}

func (reg *Registry) GetOffset(key string) (int64, error) {
	var value string

	err := reg.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(reg.bucketName)
		if bucket == nil {
			return xerrors.Errorf("%w: bucket %q not found", reg.bucketName, ErrGet)
		}

		val := bucket.Get([]byte(key))
		value = string(val)

		return nil
	})

	offset, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return 0, xerrors.Errorf("%w: %s", err.Error(), ErrGet)
	}

	return offset, nil
}

func (reg *Registry) UpdateOffset(skey string, offset int64) error {
	key := []byte(skey)
	value := []byte(strconv.FormatInt(offset, 10))

	err := reg.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(reg.bucketName)
		if err != nil {
			return xerrors.Errorf("%w: %s", err.Error(), ErrUpdate)
		}

		err = bucket.Put(key, value)
		if err != nil {
			return xerrors.Errorf("%w: %s", err.Error(), ErrUpdate)
		}
		return nil
	})

	return err
}
