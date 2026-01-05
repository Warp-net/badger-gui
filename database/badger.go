package database

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/dgraph-io/badger/v4/options"

	dsq "github.com/ipfs/go-datastore/query"
)

const (
	defaultDiscardRatioGC = 0.5
	defaultIntervalGC     = time.Hour
	defaultSleepGC        = time.Second
	defaultLimit          = 20

	ErrNotRunning    = DBError("DB is not running")
	ErrWrongPassword = DBError("wrong username or password")
)

type Key = string

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
	return nil
}

func (db *DB) IsRunning() bool {
	return db.isRunning.Load()
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
		return txn.Delete([]byte(key))
	})
}

func (db *DB) List(limit *int, startCursor *string) (keys []Key, cursor string, err error) {
	if db == nil {
		return nil, "", ErrNotRunning
	}
	if !db.isRunning.Load() {
		return nil, "", ErrNotRunning
	}
	var (
		count   = 0
		lastKey string
	)

	err = db.badger.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		opts.PrefetchValues = false

		it := txn.NewIterator(opts)
		defer it.Close()

		if startCursor != nil && *startCursor != "" {
			it.Seek([]byte(*startCursor))
			if it.Valid() && string(it.Item().Key()) == *startCursor {
				it.Next()
			}
		} else {
			it.Rewind()
		}

		for ; it.Valid(); it.Next() {
			item := it.Item()
			key := string(item.Key())

			keys = append(keys, key)
			lastKey = key
			count++

			if limit != nil && count >= *limit {
				break
			}
		}
		return nil
	})

	if limit != nil && len(keys) < *limit {
		lastKey = "end"
	}

	return keys, lastKey, err
}

func (db *DB) Search(prefix string, limit *int, offset int) (keys []Key, err error) {
	if db == nil {
		return nil, ErrNotRunning
	}
	if !db.isRunning.Load() {
		return nil, ErrNotRunning
	}
	if limit == nil {
		limit = func(i int) *int { return &i }(defaultLimit)
	}

	tx := db.badger.NewTransaction(false)
	results, err := db.query(tx, dsq.Query{
		Prefix:            prefix,
		Limit:             *limit,
		Offset:            offset,
		KeysOnly:          true,
		ReturnExpirations: false,
		ReturnsSizes:      false,
	})
	if err != nil {
		tx.Discard()
		return nil, err
	}
	entries, err := results.Rest()
	if err != nil {
		tx.Discard()
		return nil, err
	}
	for _, entry := range entries {
		keys = append(keys, entry.Key)
	}
	return keys, nil
}

func (db *DB) query(tx *badger.Txn, q dsq.Query) (_ dsq.Results, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = DBError("query recovered")
			err = fmt.Errorf("%w: %v", err, r)
		}
	}()

	if !db.IsRunning() {
		return nil, ErrNotRunning
	}
	opt := badger.DefaultIteratorOptions
	opt.PrefetchValues = !q.KeysOnly
	opt.Prefix = []byte(q.Prefix)

	// Handle ordering
	if len(q.Orders) > 0 {
		switch q.Orders[0].(type) {
		case dsq.OrderByKey, *dsq.OrderByKey:
		// We order by key by default.
		case dsq.OrderByKeyDescending, *dsq.OrderByKeyDescending:
			// Reverse order by key
			opt.Reverse = true
		default:
			// Ok, we have a weird order we can't handle. Let's
			// perform the _base_ query (prefix, filter, etc.), then
			// handle sort/offset/limit later.

			// Skip the stuff we can't apply.
			baseQuery := q
			baseQuery.Limit = 0
			baseQuery.Offset = 0
			baseQuery.Orders = nil

			// perform the base query.
			res, err := db.query(tx, baseQuery)
			if err != nil {
				return nil, err
			}

			res = dsq.ResultsReplaceQuery(res, q)

			naiveQuery := q
			naiveQuery.Prefix = ""
			naiveQuery.Filters = nil

			return dsq.NaiveQueryApply(naiveQuery, res), nil
		}
	}

	it := tx.NewIterator(opt)
	results := dsq.ResultsWithContext(q, func(ctx context.Context, output chan<- dsq.Result) {
		defer tx.Discard()
		defer it.Close()

		it.Rewind()

		for skipped := 0; skipped < q.Offset && it.Valid(); it.Next() {
			if !db.IsRunning() {
				return
			}

			if len(q.Filters) == 0 {
				skipped++
				continue
			}
			item := it.Item()

			matches := true
			check := func(value []byte) error {
				e := dsq.Entry{
					Key:   string(item.Key()),
					Value: value,
					Size:  int(item.ValueSize()),
				}

				if q.ReturnExpirations {
					e.Expiration = expires(item)
				}
				matches = filter(q.Filters, e)
				return nil
			}

			var err error
			if q.KeysOnly {
				err = check(nil)
			} else {
				err = item.Value(check)
			}

			if err != nil {
				select {
				case output <- dsq.Result{Error: err}:
				case <-db.stopChan:
					return
				case <-ctx.Done():
					return
				}
			}
			if !matches {
				skipped++
			}
		}

		for sent := 0; (q.Limit <= 0 || sent < q.Limit) && it.Valid(); it.Next() {
			if !db.IsRunning() {
				return
			}
			item := it.Item()
			e := dsq.Entry{Key: string(item.Key())}

			var result dsq.Result
			if !q.KeysOnly {
				b, err := item.ValueCopy(nil)
				if err != nil {
					result = dsq.Result{Error: err}
				} else {
					e.Value = b
					e.Size = len(b)
					result = dsq.Result{Entry: e}
				}
			} else {
				e.Size = int(item.ValueSize())
				result = dsq.Result{Entry: e}
			}

			if q.ReturnExpirations {
				result.Expiration = expires(item)
			}

			if result.Error == nil && filter(q.Filters, e) {
				continue
			}
			select {
			case output <- result:
				sent++
			case <-db.stopChan:
				return
			case <-ctx.Done():
				return
			}
		}
	})

	return results, nil
}

func filter(filters []dsq.Filter, entry dsq.Entry) bool {
	for _, f := range filters {
		if !f.Filter(entry) {
			return true
		}
	}
	return false
}

func expires(item *badger.Item) time.Time {
	expiresAt := item.ExpiresAt()
	if expiresAt > math.MaxInt64 {
		expiresAt--
	}
	return time.Unix(int64(expiresAt), 0) //#nosec
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
