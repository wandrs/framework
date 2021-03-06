// Copyright 2020 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"

	"go.wandrs.dev/framework/modules/log"
	"go.wandrs.dev/framework/modules/setting"
)

// ErrURLNotSupported represents url is not supported
var ErrURLNotSupported = errors.New("url method not supported")

// ErrInvalidConfiguration is called when there is invalid configuration for a storage
type ErrInvalidConfiguration struct {
	cfg interface{}
	err error
}

func (err ErrInvalidConfiguration) Error() string {
	if err.err != nil {
		return fmt.Sprintf("Invalid Configuration Argument: %v: Error: %v", err.cfg, err.err)
	}
	return fmt.Sprintf("Invalid Configuration Argument: %v", err.cfg)
}

// IsErrInvalidConfiguration checks if an error is an ErrInvalidConfiguration
func IsErrInvalidConfiguration(err error) bool {
	_, ok := err.(ErrInvalidConfiguration)
	return ok
}

// Type is a type of Storage
type Type string

// NewStorageFunc is a function that creates a storage
type NewStorageFunc func(ctx context.Context, cfg interface{}) (ObjectStorage, error)

var storageMap = map[Type]NewStorageFunc{}

// RegisterStorageType registers a provided storage type with a function to create it
func RegisterStorageType(typ Type, fn func(ctx context.Context, cfg interface{}) (ObjectStorage, error)) {
	storageMap[typ] = fn
}

// Object represents the object on the storage
type Object interface {
	io.ReadCloser
	io.Seeker
	Stat() (os.FileInfo, error)
}

// ObjectStorage represents an object storage to handle a bucket and files
type ObjectStorage interface {
	Open(path string) (Object, error)
	// Save store a object, if size is unknown set -1
	Save(path string, r io.Reader, size int64) (int64, error)
	Stat(path string) (os.FileInfo, error)
	Delete(path string) error
	URL(path, name string) (*url.URL, error)
	IterateObjects(func(path string, obj Object) error) error
}

// Copy copys a file from source ObjectStorage to dest ObjectStorage
func Copy(dstStorage ObjectStorage, dstPath string, srcStorage ObjectStorage, srcPath string) (int64, error) {
	f, err := srcStorage.Open(srcPath)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	size := int64(-1)
	fsinfo, err := f.Stat()
	if err == nil {
		size = fsinfo.Size()
	}

	return dstStorage.Save(dstPath, f, size)
}

// SaveFrom saves data to the ObjectStorage with path p from the callback
func SaveFrom(objStorage ObjectStorage, p string, callback func(w io.Writer) error) error {
	pr, pw := io.Pipe()
	defer pr.Close()
	go func() {
		defer pw.Close()
		if err := callback(pw); err != nil {
			_ = pw.CloseWithError(err)
		}
	}()

	_, err := objStorage.Save(p, pr, -1)
	return err
}

// Avatars represents user avatars storage
var Avatars ObjectStorage

// Init init the stoarge
func Init() error {
	return initAvatars()
}

// NewStorage takes a storage type and some config and returns an ObjectStorage or an error
func NewStorage(typStr string, cfg interface{}) (ObjectStorage, error) {
	if len(typStr) == 0 {
		typStr = string(LocalStorageType)
	}
	fn, ok := storageMap[Type(typStr)]
	if !ok {
		return nil, fmt.Errorf("Unsupported storage type: %s", typStr)
	}

	return fn(context.Background(), cfg)
}

func initAvatars() (err error) {
	log.Info("Initialising Avatar storage with type: %s", setting.Avatar.Storage.Type)
	Avatars, err = NewStorage(setting.Avatar.Storage.Type, &setting.Avatar.Storage)
	return
}
