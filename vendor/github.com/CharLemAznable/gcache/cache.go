/*
 * Simple caching library with expiration capabilities
 *     Copyright (c) 2012, Radu Ioan Fericean
 *                   2013-2017, Christian Muehlhaeuser <muesli@gmail.com>
 *
 *   For license see LICENSE.txt
 */

package gcache

import (
    "sync"
)

var (
    cacheExpireAfterWrite = make(map[string]*CacheTable)
    mutexExpireAfterWrite sync.RWMutex

    cacheExpireAfterAccess = make(map[string]*CacheTable)
    mutexExpireAfterAccess sync.RWMutex
)

// CacheXXX returns the existing cache table with given name/strategy or creates a new one
// if the table does not exist yet.

func CacheExpireAfterWrite(table string) *CacheTable {
    mutexExpireAfterWrite.RLock()
    t, ok := cacheExpireAfterWrite[table]
    mutexExpireAfterWrite.RUnlock()

    if !ok {
        mutexExpireAfterWrite.Lock()
        t, ok = cacheExpireAfterWrite[table]
        // Double check whether the table exists or not.
        if !ok {
            t = &CacheTable{
                name:     table,
                items:    make(map[interface{}]*CacheItem),
                strategy: ExpireAfterWrite,
            }
            cacheExpireAfterWrite[table] = t
        }
        mutexExpireAfterWrite.Unlock()
    }

    return t
}

func CacheExpireAfterAccess(table string) *CacheTable {
    mutexExpireAfterAccess.RLock()
    t, ok := cacheExpireAfterAccess[table]
    mutexExpireAfterAccess.RUnlock()

    if !ok {
        mutexExpireAfterAccess.Lock()
        t, ok = cacheExpireAfterAccess[table]
        // Double check whether the table exists or not.
        if !ok {
            t = &CacheTable{
                name:     table,
                items:    make(map[interface{}]*CacheItem),
                strategy: ExpireAfterAccess,
            }
            cacheExpireAfterAccess[table] = t
        }
        mutexExpireAfterAccess.Unlock()
    }

    return t
}
