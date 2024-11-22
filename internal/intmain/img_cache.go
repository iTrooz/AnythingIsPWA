package intmain

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

type imageEntry struct {
	data      []byte
	addedTime time.Time
}

type ImageCache struct {
	maxTotalSize      int64
	maxIndividualSize int64
	maxEntries        int64
	storageTime       time.Duration

	entries map[string]imageEntry
}

func randomKey() string {
	return strconv.Itoa(rand.Intn(1_000_000))
}

func NewImageCache(maxTotalSize, maxIndividualSize, maxEntries int64, storageTime time.Duration) *ImageCache {
	return &ImageCache{
		maxTotalSize:      maxTotalSize,
		maxIndividualSize: maxIndividualSize,
		maxEntries:        maxEntries,
		storageTime:       storageTime,
		entries:           make(map[string]imageEntry),
	}
}

type AddReturn int64

const (
	Success           AddReturn = 0
	EntrySizeExceeded AddReturn = 1
	TotalSizeExceeded AddReturn = 2
)

// return site in bytes
func (c *ImageCache) ByteSize() int {
	size := 0
	for _, v := range c.entries {
		size += len(v.data)
	}
	return size
}

func (c *ImageCache) canBeAdded(value []byte) error {
	if c.ByteSize()+len(value) > int(c.maxTotalSize) {
		return fmt.Errorf("total size exceeded")
	}
	if int64(len(value)) > c.maxIndividualSize {
		return fmt.Errorf("entry size exceeded")
	}
	if int64(len(c.entries)) >= c.maxEntries {
		return fmt.Errorf("max entries exceeded")
	}
	return nil
}

// Evict data which has passed timeout
func (c *ImageCache) Evict() {
	evicted := 0
	for k, v := range c.entries {
		if time.Since(v.addedTime) > c.storageTime {
			delete(c.entries, k)
			evicted++
		}
	}
	logrus.Infof("Evicted %d entries", evicted)
}

func (c *ImageCache) Add(value []byte) (string, error) {
	// check if can be added
	if c.canBeAdded(value) != nil {
		c.Evict()
	}
	err := c.canBeAdded(value)
	if err != nil {
		return "", err
	}

	// add
	var key string
	for {
		key = randomKey()
		if _, ok := c.entries[key]; !ok {
			break
		}

	}
	c.entries[key] = imageEntry{
		data:      value,
		addedTime: time.Now(),
	}

	logrus.Debugf("Added entry with key %s (size=%v)", key, len(value))

	return key, nil
}
