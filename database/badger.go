package database

import (
	"encoding/hex"
	"errors"
	"github.com/filinvadim/badger-gui/domain"
	"log"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"
)

/*
  BadgerDB is a high-performance, embedded key-value database
  written in Go, utilizing LSM-trees (Log-Structured Merge-Trees) for efficient
  data storage and processing. It is designed for high-load scenarios that require
  minimal latency and high throughput.

  Key Features:
    - Embedded: Operates within the application without the need for a separate database: server.
    - Key-Value Store: Enables storing and retrieving data by key using efficient indexing.
    - LSM Architecture: Provides high write speed due to log-structured data storage.
    - Zero GC Overhead: Minimizes garbage collection impact by directly working with mmap and byte slices.
    - ACID Transactions: Supports transactions with snapshot isolation.
    - In-Memory and Disk Mode: Allows storing data in RAM or on SSD/HDD.
    - Low Resource Consumption: Suitable for embedded systems and server applications with limited memory.

  BadgerDB is used in cases where:
    - High Performance is required: It is faster than traditional disk-based databases (e.g., BoltDB) due to the LSM structure.
    - Embedded Storage is needed: No need to run a separate database: server (unlike Redis, PostgreSQL, etc.).
    - Efficient Streaming Writes are required: Suitable for logs, caches, message brokers, and other write-intensive workloads.
    - Transaction Support is necessary: Allows safely executing multiple operations within a single transaction.
    - Large Data Volumes are handled: Supports sharding and disk offloading, useful for processing massive datasets.
    - Flexibility is key: Easily integrates into distributed systems and P2P applications.

  BadgerDB is especially useful for systems where high write speed, low overhead, and the ability to operate without an external database: server are critical.
  https://github.com/dgraph-io/badger
*/

const (
	defaultDiscardRatioGC = 0.5
	defaultIntervalGC     = time.Hour
	defaultSleepGC        = time.Second

	ErrNotRunning    = DBError("DB is not running")
	ErrWrongPassword = DBError("wrong username or password")
)

type DBError string

func (e DBError) Error() string {
	return string(e)
}

type Options struct {
	discardRatioGC float64
	intervalGC     time.Duration
	sleepGC        time.Duration
}

type DB struct {
	badger *badger.DB

	isRunning *atomic.Bool

	badgerOpts     badger.Options
	discardRatioGC float64
	intervalGC     time.Duration
	sleepGC        time.Duration

	stopChan chan struct{}
}

func New(o *Options) (*DB, error) {
	if o == nil {
		o = &Options{}
	}

	currentDir, _ := os.Getwd()
	badgerOpts := badger.
		DefaultOptions(currentDir + "/.badger").
		WithSyncWrites(true).
		WithIndexCacheSize(256 << 20).
		WithCompression(options.None).
		WithNumCompactors(4).
		WithLoggingLevel(badger.ERROR).
		WithBlockCacheSize(512 << 20)

	if o.intervalGC == 0 {
		o.intervalGC = defaultIntervalGC
	}
	if o.discardRatioGC == 0 {
		o.discardRatioGC = defaultDiscardRatioGC
	}
	if o.sleepGC == 0 {
		o.sleepGC = defaultSleepGC
	}

	storage := &DB{
		badger: nil, stopChan: make(chan struct{}), isRunning: new(atomic.Bool),
		badgerOpts:     badgerOpts,
		discardRatioGC: o.discardRatioGC, intervalGC: o.intervalGC, sleepGC: o.sleepGC,
	}

	return storage, nil
}

func (db *DB) Open(dbPath, key, compression string) (err error) {
	if dbPath == "" {
		db.badgerOpts = badger.
			DefaultOptions("").
			WithDir("").
			WithValueDir("").
			WithInMemory(true).
			WithSyncWrites(true).
			WithIndexCacheSize(256 << 20).
			WithNumCompactors(2).
			WithLoggingLevel(badger.ERROR).
			WithBlockCacheSize(512 << 20)
	} else {
		db.badgerOpts = db.badgerOpts.WithDir(dbPath)
	}
	if key != "" {
		if hexKey, err := hex.DecodeString(key); err == nil {
			key = string(hexKey)
		}
		db.badgerOpts = db.badgerOpts.WithEncryptionKey([]byte(key))
	}
	if compression != "" {
		switch strings.ToLower(compression) {
		case "snappy":
			db.badgerOpts = db.badgerOpts.WithCompression(options.Snappy)
		case "zstd":
			db.badgerOpts = db.badgerOpts.WithCompression(options.ZSTD)
		default:
			db.badgerOpts = db.badgerOpts.WithCompression(options.None)
		}
	}

	db.badger, err = badger.Open(db.badgerOpts)
	if errors.Is(err, badger.ErrEncryptionKeyMismatch) {
		return ErrWrongPassword
	}
	if err != nil {
		return err
	}
	db.isRunning.Store(true)
	if !db.badgerOpts.InMemory {
		go db.runEventualGC()
	}

	return nil
}

func (db *DB) IsRunning() bool {
	return db.isRunning.Load()
}

func (db *DB) runEventualGC() {
	if db.badgerOpts.InMemory {
		return
	}
	log.Println("database: garbage collection started")
	gcTicker := time.NewTicker(db.intervalGC)
	defer gcTicker.Stop()

	_ = db.badger.RunValueLogGC(db.discardRatioGC)
	for {
		select {
		case <-gcTicker.C:
			for {
				err := db.badger.RunValueLogGC(db.discardRatioGC)
				if errors.Is(err, badger.ErrNoRewrite) ||
					errors.Is(err, badger.ErrRejected) {
					break
				}
				time.Sleep(db.sleepGC)
			}
			log.Println("database: garbage collection complete")
		case <-db.stopChan:
			return
		}
	}
}

func (db *DB) Set(key string, value []byte) error {
	if db == nil {
		return ErrNotRunning
	}
	if !db.isRunning.Load() {
		return ErrNotRunning
	}

	return db.badger.Update(func(txn *badger.Txn) error {
		e := badger.NewEntry([]byte(key), value)
		return txn.SetEntry(e)
	})
}

func (db *DB) Get(key string) ([]byte, error) {
	if db == nil {
		return nil, ErrNotRunning
	}
	if !db.isRunning.Load() {
		return nil, ErrNotRunning
	}

	var result []byte
	err := db.badger.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		result = append([]byte{}, val...)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (db *DB) Delete(key string) error {
	if db == nil {
		return ErrNotRunning
	}
	if !db.isRunning.Load() {
		return ErrNotRunning
	}

	return db.badger.Update(func(txn *badger.Txn) error {
		if err := txn.Delete([]byte(key)); err != nil {
			return err
		}
		return txn.Delete([]byte(key))
	})
}

const endCursor = "end"

func (db *DB) List(prefix string, limit *int, cursor *string) (domain.Items, string, error) {
	var startCursor string
	if cursor != nil && *cursor != "" {
		startCursor = *cursor
	}
	if startCursor == endCursor {
		return []domain.Item{}, endCursor, nil
	}

	if limit == nil {
		defaultLimit := 20
		limit = &defaultLimit
	}

	items := make([]domain.Item, 0, *limit) //
	cur, err := db.iterate(
		prefix, startCursor, *limit,
		func(key string, value []byte) {
			items = append(items, domain.Item{
				Key:   key,
				Value: string(value),
			})
			return
		},
	)
	return items, cur, err
}

type iterKeysValuesFunc func(key string, val []byte)

func (db *DB) iterate(
	prefix string,
	startCursor string,
	limit int,
	handler iterKeysValuesFunc,
) (cursor string, err error) {
	if startCursor == endCursor {
		return endCursor, nil
	}
	opts := badger.DefaultIteratorOptions
	opts.PrefetchValues = false
	opts.PrefetchSize = limit * 2

	txn := db.badger.NewTransaction(false)
	defer txn.Discard()
	it := txn.NewIterator(opts)

	var (
		lastKey string
		iterNum int
		errs    []error
	)

	seekKey := []byte(prefix)
	if startCursor != "" {
		seekKey = []byte(startCursor)
	}

	for it.Seek(seekKey); it.ValidForPrefix([]byte(prefix)); it.Next() {
		item := it.Item()
		key := string(item.Key())

		if iterNum >= limit {
			lastKey = key
			break
		}

		val, err := item.ValueCopy(nil)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		handler(key, val)
		iterNum++
	}

	if iterNum < limit {
		lastKey = endCursor
	}
	it.Close()
	if err := txn.Commit(); err != nil {
		return "", err
	}
	return lastKey, errors.Join(errs...)
}

func (db *DB) Close() {
	if db == nil {
		return
	}
	log.Println("closing database...")

	if db.badger == nil {
		return
	}
	if !db.isRunning.Load() {
		return
	}
	close(db.stopChan)

	if err := db.badger.Close(); err != nil {
		log.Printf("database: close: %v", err)
		return
	}
	db.isRunning.Store(false)
	db.badger = nil
}
