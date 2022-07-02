/*
Copyright The Helm Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cache

import (
	"context"
	"time"

	"github.com/bsm/redislock"
	"github.com/go-redis/redis/v8"
)

type (
	// RedisStore implements the Store interface, used for storing objects in-memory
	RedisStore struct {
		Client     *redis.Client
		LockClient *redislock.Client
	}
)

// NewRedisStore creates a new RedisStore
func NewRedisStore(addr string, password string, db int) *RedisStore {
	store := &RedisStore{}
	redisClientOptions := &redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	}
	store.Client = redis.NewClient(redisClientOptions)
	store.LockClient = redislock.New(store.Client)
	return store
}

// Get returns an object at key
func (store *RedisStore) Get(key string) ([]byte, error) {
	// TODO: pass context in?
	content, err := store.Client.Get(context.Background(), key).Bytes()
	return content, err
}

// Set saves a new value for key
func (store *RedisStore) Set(key string, contents []byte) error {
	// TODO: pass context in?
	err := store.Client.Set(context.Background(), key, contents, 0).Err()
	return err
}

// Delete removes a key from the store
func (store *RedisStore) Delete(key string) error {
	// TODO: pass context in
	err := store.Client.Del(context.Background(), key).Err()
	return err
}

// Lock returns a lock on a key
func (store *RedisStore) Lock(key string) (*redislock.Lock, error) {
	// TODO: pass context in
	// Retry every 100ms, for up-to 100x (10 seconds total)
	backoff := redislock.LimitRetry(redislock.LinearBackoff(100*time.Millisecond), 100)
	lock, err := store.LockClient.Obtain(context.Background(), key, time.Second*30, &redislock.Options{
		RetryStrategy: backoff,
	})
	if err != nil {
		return nil, err
	}
	return lock, nil
}
