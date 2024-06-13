// Copyright (c) 2022-2024 The Focela Authors, All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package redis provides convenient client for redis server.
//
// Redis Client.
//
// Redis Commands Official: https://redis.io/commands
//
// Redis Chinese Documentation: http://redisdoc.com/
package redis

import (
	"context"

	"github.com/gocarp/codes"
	"github.com/gocarp/errors"
	"github.com/gocarp/go/container/vars"
	"github.com/gocarp/helpers/text/str"
)

// Redis client.
type Redis struct {
	config *Config
	localAdapter
	localGroup
}

type (
	localGroup struct {
		localGroupGeneric
		localGroupHash
		localGroupList
		localGroupPubSub
		localGroupScript
		localGroupSet
		localGroupSortedSet
		localGroupString
	}
	localAdapter        = Adapter
	localGroupGeneric   = GroupGenericInterface
	localGroupHash      = GroupHashInterface
	localGroupList      = GroupListInterface
	localGroupPubSub    = GroupPubSubInterface
	localGroupScript    = GroupScriptInterface
	localGroupSet       = GroupSetInterface
	localGroupSortedSet = GroupSortedSetInterface
	localGroupString    = GroupStringInterface
)

const (
	errorNilRedis = `the Redis object is nil`
)

var (
	errorNilAdapter = str.Trim(str.Replace(`
redis adapter is not set, missing configuration or adapter register? 
possible reference: https://github.com/gocarp/contrib/tree/master/nosql/redis
`, "\n", ""))
)

// AdapterFunc is the function creating redis adapter.
type AdapterFunc func(config *Config) Adapter

var (
	// defaultAdapterFunc is the default adapter function creating redis adapter.
	defaultAdapterFunc AdapterFunc = func(config *Config) Adapter {
		return nil
	}
)

// initGroup initializes the group object of redis.
func (r *Redis) initGroup() *Redis {
	r.localGroup = localGroup{
		localGroupGeneric:   r.localAdapter.GroupGeneric(),
		localGroupHash:      r.localAdapter.GroupHash(),
		localGroupList:      r.localAdapter.GroupList(),
		localGroupPubSub:    r.localAdapter.GroupPubSub(),
		localGroupScript:    r.localAdapter.GroupScript(),
		localGroupSet:       r.localAdapter.GroupSet(),
		localGroupSortedSet: r.localAdapter.GroupSortedSet(),
		localGroupString:    r.localAdapter.GroupString(),
	}
	return r
}

// SetAdapter changes the underlying adapter with custom adapter for current redis client.
func (r *Redis) SetAdapter(adapter Adapter) {
	if r == nil {
		panic(errors.NewCode(codes.CodeInvalidParameter, errorNilRedis))
	}
	r.localAdapter = adapter
}

// GetAdapter returns the adapter that is set in current redis client.
func (r *Redis) GetAdapter() Adapter {
	if r == nil {
		return nil
	}
	return r.localAdapter
}

// Conn retrieves and returns a connection object for continuous operations.
// Note that you should call Close function manually if you do not use this connection any further.
func (r *Redis) Conn(ctx context.Context) (Conn, error) {
	if r == nil {
		return nil, errors.NewCode(codes.CodeInvalidParameter, errorNilRedis)
	}
	if r.localAdapter == nil {
		return nil, errors.NewCode(codes.CodeNecessaryPackageNotImport, errorNilAdapter)
	}
	return r.localAdapter.Conn(ctx)
}

// Do send a command to the server and returns the received reply.
// It uses json.Marshal for struct/slice/map type values before committing them to redis.
func (r *Redis) Do(ctx context.Context, command string, args ...interface{}) (*vars.Var, error) {
	if r == nil {
		return nil, errors.NewCode(codes.CodeInvalidParameter, errorNilRedis)
	}
	if r.localAdapter == nil {
		return nil, errors.NewCodef(codes.CodeMissingConfiguration, errorNilAdapter)
	}
	return r.localAdapter.Do(ctx, command, args...)
}

// MustConn performs as function Conn, but it panics if any error occurs internally.
func (r *Redis) MustConn(ctx context.Context) Conn {
	c, err := r.Conn(ctx)
	if err != nil {
		panic(err)
	}
	return c
}

// MustDo performs as function Do, but it panics if any error occurs internally.
func (r *Redis) MustDo(ctx context.Context, command string, args ...interface{}) *vars.Var {
	v, err := r.Do(ctx, command, args...)
	if err != nil {
		panic(err)
	}
	return v
}

// Close closes current redis client, closes its connection pool and releases all its related resources.
func (r *Redis) Close(ctx context.Context) error {
	if r == nil || r.localAdapter == nil {
		return nil
	}
	return r.localAdapter.Close(ctx)
}

// New creates and returns a redis client.
// It creates a default redis adapter of go-redis.
func New(config ...*Config) (*Redis, error) {
	var (
		usedConfig  *Config
		usedAdapter Adapter
	)
	if len(config) > 0 && config[0] != nil {
		// Redis client with go redis implements adapter from given configuration.
		usedConfig = config[0]
		usedAdapter = defaultAdapterFunc(config[0])
	} else if configFromGlobal, ok := GetConfig(); ok {
		// Redis client with go redis implements adapter from package configuration.
		usedConfig = configFromGlobal
		usedAdapter = defaultAdapterFunc(configFromGlobal)
	}
	if usedConfig == nil {
		return nil, errors.NewCode(
			codes.CodeInvalidConfiguration,
			`no configuration found for creating Redis client`,
		)
	}
	if usedAdapter == nil {
		return nil, errors.NewCode(
			codes.CodeNecessaryPackageNotImport,
			errorNilAdapter,
		)
	}
	redis := &Redis{
		config:       usedConfig,
		localAdapter: usedAdapter,
	}
	return redis.initGroup(), nil
}

// NewWithAdapter creates and returns a redis client with given adapter.
func NewWithAdapter(adapter Adapter) (*Redis, error) {
	if adapter == nil {
		return nil, errors.NewCodef(codes.CodeInvalidParameter, `adapter cannot be nil`)
	}
	redis := &Redis{localAdapter: adapter}
	return redis.initGroup(), nil
}

// RegisterAdapterFunc registers default function creating redis adapter.
func RegisterAdapterFunc(adapterFunc AdapterFunc) {
	defaultAdapterFunc = adapterFunc
}
