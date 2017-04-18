package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"bytes"
	"os"
	"strings"
	"github.com/peterbourgon/diskv"
	"crypto/sha256"
	"encoding/base64"
	"encoding/gob"
	"time"
	"path/filepath"
	"reflect"
	"syscall"
)

type Request struct {
	Type      string `json:"type"`
	MBean     string `json:"mbean"`
	Attribute string `json:"attribute,omitempty"`
	Path      string `json:"path,omitempty"`
	Config    map[string]interface{} `json:"config,omitempty"`
}

type Response struct {
	Request   Request
	Value     interface{}
	Timestamp int64
	Error     string
	Status    int
}

type CacheEntry struct {
	Path         string
	// when it was last fetched from JMX
	LastFetch    int64
	// when it was last read by Zabbix
	LastRead     int64
	// last interval between reads
	LastInterval int64
	CachedValue  interface{}
}

func NewRequest(path string) Request {
	parts := strings.SplitN(path, "/", 3)

	ret := Request{
		Type: "read",
		MBean: parts[0],
		Config: map[string]interface{}{
			"includeStackTrace": false,
		},
	}
	if len(parts) > 1 {
		ret.Attribute = parts[1]
	}
	if len(parts) > 2 {
		ret.Path = parts[2]
	}
	return ret
}

func RequestPath(request Request) string {
	var path = request.MBean
	if len(request.Attribute) > 0 {
		path += "/" + request.Attribute
		if len(request.Path) > 0 {
			path += "/" + request.Path
		}
	}
	return path
}

func JolokiaRead(url string, paths []string) (ret map[string]interface{}) {
	ret = make(map[string]interface{})

	requests := make([]Request, len(paths))
	for i, path := range paths {
		requests[i] = NewRequest(path)
	}

	payload, err := json.Marshal(requests)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	client := &http.Client{Timeout:3 * time.Second}
	response, err := client.Post(url, "application/json", bytes.NewReader(payload))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	if response.StatusCode != http.StatusOK {
		fmt.Fprintln(os.Stderr, "Unexpected HTTP response status", response.StatusCode)
		return
	}

	var results []Response
	decoder := json.NewDecoder(response.Body)
	decoder.UseNumber()
	err = decoder.Decode(&results)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to parse Jolokia response", err)
		return
	}

	for _, resp := range results {
		path := NormalizePath(RequestPath(resp.Request))
		if resp.Status == http.StatusOK {
			value := resp.Value
			if value != nil && reflect.TypeOf(value).Name() == "Number" {
				value = value.(json.Number).String()
			}
			ret[path] = value
		} else {
			fmt.Fprintf(os.Stderr, "Could not read %s: %s\n", path, resp.Error)
		}
	}
	return
}

func KeyHash(value string) string {
	hash := sha256.New()
	hash.Write([]byte(value))
	return base64.URLEncoding.EncodeToString(hash.Sum(nil))
}

func ReadCache(disk *diskv.Diskv, path string) CacheEntry {
	data, err := disk.Read(KeyHash(path))
	if err != nil {
		return CacheEntry{Path:path}
	}

	var entry CacheEntry
	err = gob.NewDecoder(bytes.NewReader(data)).Decode(&entry)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to parse cache entry", path, err)
		return CacheEntry{Path:path}
	}
	return entry
}

func ReadCacheByKey(disk *diskv.Diskv, key string) CacheEntry {
	data, err := disk.Read(key)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return CacheEntry{}
	}

	var entry CacheEntry
	err = gob.NewDecoder(bytes.NewReader(data)).Decode(&entry)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to parse cache entry", key, err)
		return CacheEntry{}
	}
	return entry
}

func WriteCache(disk *diskv.Diskv, entry *CacheEntry) {
	var buf bytes.Buffer
	err := gob.NewEncoder(&buf).Encode(entry)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to serialize cache entry", entry, err)
		return
	}

	err = disk.Write(KeyHash(entry.Path), buf.Bytes())
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to store cache entry", entry, err)
		os.Exit(1)
	}
}

func main() {
	if len(os.Args) < 4 {
		fmt.Fprintln(os.Stderr, "Too few arguments")
		os.Exit(1)
	}

	basedir := os.Args[1]
	url := os.Args[2]
	path := NormalizePath(ZabbixUnescape(strings.Join(NonEmpty(os.Args[3:]), ",")))
	cachedir := filepath.Join(basedir, KeyHash(url))

	// obtain exclusive lock
	lock, err := os.Create(cachedir + ".lock")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to create lock file", cachedir + ".lock", err)
		os.Exit(1)
	}
	syscall.Flock(int(lock.Fd()), syscall.LOCK_EX)

	disk := diskv.New(diskv.Options{
		BasePath:cachedir,
		CacheSizeMax: 1024 * 1024,
	})
	entry := ReadCache(disk, path)

	now := time.Now().Unix()
	if entry.CachedValue != nil && entry.LastFetch > now - entry.LastInterval * 2 {
		// entry has a recently cached value
		value := entry.CachedValue
		if entry.LastRead != 0 {
			entry.LastInterval = now - entry.LastRead
		}
		entry.LastRead = now
		entry.CachedValue = nil
		WriteCache(disk, &entry)
		fmt.Print(value)
	} else if entry.LastInterval > 0 && entry.LastInterval < 400 && now > entry.LastFetch + entry.LastInterval / 2 {
		// regularly fetched item - fetch other pending values too
		entries := []*CacheEntry{&entry}
		paths := []string{entry.Path}
		for key := range disk.Keys(nil) {
			e := ReadCacheByKey(disk, key)
			if e.CachedValue != nil || e.LastInterval == 0 || e.LastInterval >= 400 {
				continue
			}
			if e.Path == entry.Path {
				// requested item will be processed anyway
				continue
			}
			if e.LastRead < now - 600 {
				// remove unused entries
				disk.Erase(key)
				continue
			}
			if now > entry.LastFetch + entry.LastInterval / 2 {
				entries = append(entries, &e)
				paths = append(paths, e.Path)
			}
		}
		values := JolokiaRead(url, paths)

		now = time.Now().Unix()
		for _, e := range entries {
			value := values[e.Path]
			e.LastFetch = now
			e.CachedValue = value
			if e != &entry {
				WriteCache(disk, e)
			}
		}

		value := values[entry.Path]
		entry.LastFetch = now
		if entry.LastRead != 0 {
			entry.LastInterval = now - entry.LastRead
		}
		entry.LastRead = now
		entry.CachedValue = nil
		WriteCache(disk, &entry)
		if value != nil {
			fmt.Print(value)
		} else {
			os.Exit(1)
		}
	} else {
		// fetch item individually
		value := JolokiaRead(url, []string{path})[entry.Path]

		now = time.Now().Unix()
		entry.LastFetch = now
		if entry.LastRead != 0 {
			entry.LastInterval = now - entry.LastRead
		}
		entry.LastRead = now
		entry.CachedValue = nil
		WriteCache(disk, &entry)
		if value != nil {
			fmt.Print(value)
		} else {
			os.Exit(1)
		}
	}
}
