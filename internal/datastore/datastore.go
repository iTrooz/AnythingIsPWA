package datastore

import "time"

type DataStoreEntry struct {
	value      []byte
	storedDate time.Time
}

type DataStore struct {
	maxSize     int64
	maxValues   int
	data        map[string]DataStoreEntry
	currentSize int64
	expiration  time.Duration
}

func NewDataStore(maxSize int64, maxValues int, expiration time.Duration) DataStore {
	return DataStore{
		data:        make(map[string]DataStoreEntry),
		currentSize: 0,

		maxSize:    maxSize,
		maxValues:  maxValues,
		expiration: expiration,
	}
}

func (ds *DataStore) canInsert(value []byte) bool {
	return ds.currentSize+int64(len(value)) <= ds.maxSize && len(ds.data) < ds.maxValues
}

func (ds *DataStore) Set(key string, value []byte) bool {
	if !ds.canInsert(value) {
		ds.ExpirationCheck()
		if !ds.canInsert(value) {
			return false
		}
	}

	ds.data[key] = DataStoreEntry{
		value:      value,
		storedDate: time.Now(),
	}
	ds.currentSize += int64(len(value))
	return true
}

func (ds *DataStore) Get(key string) []byte {
	value, ok := ds.data[key]
	if ok {
		return value.value
	}
	return nil
}

func (ds *DataStore) ExpirationCheck() int {
	n := 0
	for key, value := range ds.data {
		if time.Since(value.storedDate) > ds.expiration {
			delete(ds.data, key)
			ds.currentSize -= int64(len(value.value))
			n++
		}
	}
	return n
}
