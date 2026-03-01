package driver

import (
	"sync"

	"github.com/dgraph-io/badger/v4"
	"github.com/leeseika/cv-demo/pkg/config"
)

var (
	_badgerDB         *badger.DB
	_initBadgerDBOnce sync.Once
)

func InitBadgerDB(conf config.BadgerDB) {
	_initBadgerDBOnce.Do(func() {
		opts := badger.DefaultOptions(conf.Path)
		db, err := badger.Open(opts)
		if err != nil {
			panic("failed to open BadgerDB: " + err.Error())
		}
		_badgerDB = db
	})
}

func GetBadgerDB() *badger.DB {
	return _badgerDB
}
