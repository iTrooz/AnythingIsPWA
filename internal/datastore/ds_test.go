package datastore

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSetGet(t *testing.T) {
	ds := NewDataStore(1000*1000, 1, time.Hour)
	assert.True(t, ds.Set("key1", []byte("value")))
	assert.Equal(t, []byte("value"), ds.Get("key1"))
}

func TestInvalidKey(t *testing.T) {
	ds := NewDataStore(1000*1000, 1, time.Hour)
	assert.Equal(t, []byte(nil), ds.Get("key1"))
	assert.True(t, ds.Get("key1") == nil)
}

func TestMaxValues(t *testing.T) {
	ds := NewDataStore(1000*1000, 1, time.Hour)
	assert.True(t, ds.Set("key1", []byte("value")))
	assert.False(t, ds.Set("key2", []byte("value")))
}

func TestMaxSize(t *testing.T) {
	ds := NewDataStore(2, 1, time.Hour)
	assert.False(t, ds.Set("key1", []byte("123")))
	assert.True(t, ds.Set("key2", []byte("12")))
	assert.False(t, ds.Set("key3", []byte("1")))
}

func TestExpiration(t *testing.T) {
	ds := NewDataStore(1000*1000, 1, time.Second)
	assert.True(t, ds.Set("key1", []byte("value")))
	ds.ExpirationCheck()
	assert.Equal(t, []byte("value"), ds.Get("key1"))
	
	time.Sleep(1*time.Second)
	ds.ExpirationCheck()
	assert.Equal(t, []byte(nil), ds.Get("key1"))
}
