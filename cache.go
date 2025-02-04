package main

import (
	"container/list"
	"io/fs"
	"sync"
)

const (
	defaultCacheSize = 50 * 1024 * 1024
)

type CacheFile struct {
	Contents []byte
	FileInfo fs.FileInfo
}

type cacheItem struct {
	Key   string
	Value *CacheFile
}

type MRUCache struct {
	capacity int64
	size     int64
	items    sync.Map // string -> *list.Element
	order    *list.List
}

func NewMRUCache(capacity int64) *MRUCache {
	if capacity <= 0 {
		capacity = defaultCacheSize
	}

	return &MRUCache{
		capacity: capacity,
		size:     0,
		order:    list.New(),
	}
}

func (cache *MRUCache) Put(key string, value *CacheFile) {
	if value.FileInfo.Size() > cache.capacity {
		return
	}

	if element, exists := cache.items.Load(key); exists {
		cache.order.MoveToFront(element.(*list.Element))
		cache.evictIfNeeded()

		return
	}

	item := &cacheItem{Key: key, Value: value}
	element := cache.order.PushFront(item)
	cache.items.Store(key, element)
	cache.size += value.FileInfo.Size()

	cache.evictIfNeeded()
}

func (cache *MRUCache) Exists(key string) bool {
	_, exists := cache.items.Load(key)

	return exists
}

func (cache *MRUCache) Get(key string) (*CacheFile, bool) {
	if element, exists := cache.items.Load(key); exists {
		cache.order.MoveToFront(element.(*list.Element))

		return element.(*list.Element).Value.(*cacheItem).Value, true
	}

	return nil, false
}

func (cache *MRUCache) Open(key string) (fs.File, error) {
	if file, exists := cache.Get(key); exists {
		return NewVirtualFile(file.Contents, file.FileInfo), nil
	}

	return nil, fs.ErrNotExist
}

func (cache *MRUCache) evictIfNeeded() {
	for cache.size > cache.capacity {
		element := cache.order.Back()
		if element == nil {
			break
		}
		item := element.Value.(*cacheItem)
		cache.size -= item.Value.FileInfo.Size()
		cache.items.Delete(item.Key)
		cache.order.Remove(element)
	}
}

type VirtualFile struct {
	Contents []byte
	FileInfo fs.FileInfo
}

func NewVirtualFile(contents []byte, fileInfo fs.FileInfo) *VirtualFile {
	return &VirtualFile{
		Contents: contents,
		FileInfo: fileInfo,
	}
}

func (f *VirtualFile) Read(p []byte) (n int, err error) {
	return copy(p, f.Contents), nil
}

func (f *VirtualFile) Close() error {
	return nil
}

func (f *VirtualFile) Stat() (fs.FileInfo, error) {
	return f.FileInfo, nil
}
