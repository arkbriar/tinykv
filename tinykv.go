//  Copyright © 2018 The TinyKV Authors.
//
//  Permission is hereby granted, free of charge, to any person obtaining a copy
//  of this software and associated documentation files (the "Software"), to deal
//  in the Software without restriction, including without limitation the rights
//  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
//  copies of the Software, and to permit persons to whom the Software is
//  furnished to do so, subject to the following conditions:
//
//  The above copyright notice and this permission notice shall be included in
//  all copies or substantial portions of the Software.
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
//  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
//  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
//  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
//  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
//  THE SOFTWARE.

package tinykv

import (
	"fmt"
)

// Byte constants to make codes more understandable.
const (
	KiloByte = int64(1) << 10
	MegaByte = KiloByte << 10
	GigaByte = MegaByte << 10
	TeraByte = GigaByte << 10
	PetaByte = TeraByte << 10
)

/*******************************************************************************
  Here we define DBError which should be used over this project in order to
  distinguish different errors.
 ******************************************************************************/

// ErrorCode represents the error types.
type ErrorCode int

// ErrorCode constants including NotFound, Corruption, NotSupported, InvalidArgument
// and IOError.
const (
	NotFound        ErrorCode = 0x0
	Corruption      ErrorCode = 0x1
	NotSupported    ErrorCode = 0x2
	InvalidArgument ErrorCode = 0x3
	IOError         ErrorCode = 0x4
)

// DBError is error type used in TinyKV project.
type DBError struct {
	Code   ErrorCode
	Status string
}

// Implements error interface
func (err *DBError) Error() string {
	return fmt.Sprintf("error code: %d, details: %s", err.Code, err.Status)
}

// IsNotFoundError tells if the error is a DBError and is a not-found error.
func IsNotFoundError(err error) bool {
	if dbErr, ok := err.(*DBError); ok {
		return dbErr.Code == NotFound
	}
	return false
}

// IsCorruptionError tells if the error is a DBError and is a corruption error.
func IsCorruptionError(err error) bool {
	if dbErr, ok := err.(*DBError); ok {
		return dbErr.Code == Corruption
	}
	return false
}

// IsNotSupportedError tells if the error is a DBError and is a not-supported error.
func IsNotSupportedError(err error) bool {
	if dbErr, ok := err.(*DBError); ok {
		return dbErr.Code == NotSupported
	}
	return false
}

// IsInvalidArgumentError tells if the error is a DBError and is a invalid-argument error.
func IsInvalidArgumentError(err error) bool {
	if dbErr, ok := err.(*DBError); ok {
		return dbErr.Code == InvalidArgument
	}
	return false
}

// IsIOError tells if the error is a DBError and is a I/O error.
func IsIOError(err error) bool {
	if dbErr, ok := err.(*DBError); ok {
		return dbErr.Code == IOError
	}
	return false
}

/*******************************************************************************
  Here we define Options, WriteOptions and ReadOptions to be used in DB.
 ******************************************************************************/

// CompressionType represents the compression types.
// Currently only no compression and compression using snappy are supported.
type CompressionType int

// CompressionType constants.
const (
	NoCompression     CompressionType = 0x0
	SnappyCompression CompressionType = 0x1
)

// Options to control the behavior of a database.
type Options struct {
	CreateIfMissing bool
	ErrorIfExists   bool
	ParanoidChecks  bool
	WriteBufferSize int64
	MaxOpenFiles    int
	Compression     CompressionType
}

// ReadOptions to control read operations.
type ReadOptions struct {
	VerifyChecksum bool
	FillCache      bool
}

// WriteOptions to control write operations.
type WriteOptions struct {
	Sync bool
}

// NewOptions creates a default Options.
func NewOptions() Options {
	return Options{
		CreateIfMissing: true,
		ErrorIfExists:   true,
		ParanoidChecks:  false,
		WriteBufferSize: MegaByte * 4,
		MaxOpenFiles:    1000,
		Compression:     SnappyCompression,
	}
}

// NewReadOptions creates a default ReadOptions.
func NewReadOptions() ReadOptions {
	return ReadOptions{
		VerifyChecksum: false,
		FillCache:      false,
	}
}

// NewWriteOptions creates a default WriteOptions.
func NewWriteOptions() WriteOptions {
	return WriteOptions{
		Sync: false,
	}
}

/*******************************************************************************
  Here we define the DB interface and kvDB struct which implements DB.
 ******************************************************************************/

// DB is a persistent ordered map from keys to values.
// DB is safe for concurrent access from multiple threads without
// any external synchronization.
type DB interface {
	// Open the database. Returns nil on success, and an error otherwise.
	Open() error

	// Close the database. Returns nil on success, and an error otherwise.
	// Operations on a closed database will get a not supported error. You can distinguish it
	// using IsNotSupportedError(error).
	Close() error

	// Put key-value pair into the database. Returns nil on success, and an error otherwise.
	// If key is empty, it returns a invalid argument error. You can distinguish it using
	// IsInvalidArgumentError(error).
	Put(key, value string, options WriteOptions) error

	// Perform a read-modify-write atomic operation on database entry for key.
	// Returns old value and nil on success, and nil and an error otherwise.
	// If there's no such key, the old value returned should be empty.
	// If key is empty, it returns a invalid argument error (and can be distinguished by
	// IsInvalidArgumentError(error)).
	ReadModifyWrite(key, value string, options WriteOptions) (string, error)

	// Get value of specified key from the database. Returns value and nil on success,
	// and nil and an error otherwise.
	// If key is not found, you will get an not found error. You can distinguish it
	// using IsNotFoundError(error).
	Get(key string, options ReadOptions) (string, error)

	// Delete the database entry for key. Returns nil on success, and an error otherwise.
	// It is not an error if key does not exists in the database.
	Delete(key string, options WriteOptions) error
}

// kvDB is the default implementation of DB interface.
type kvDB struct {
	// force implements DB interface
	DB

	// database name
	name string
	// database options
	options Options
}

// NewDB creates a new database instance with "name" and given options.
func NewDB(name string, options Options) DB {
	return &kvDB{options: options, name: name}
}

// DestroyDB destroys the database with "name" and given options.
// Returns nil on success and an error otherwise.
func DestroyDB(name string, options Options) error {
	// TODO
	return &DBError{NotSupported, "unimplemented"}
}

// RepairDB tries to repair the database with "name" and given options.
// Returns nil on success and an error otherwise.
func RepairDB(name string, options Options) error {
	// TODO
	return &DBError{NotSupported, "unimplemented"}
}

func (db *kvDB) Open() error {
	// TODO
	return &DBError{NotSupported, "unimplemented"}
}

func (db *kvDB) Close() error {
	// TODO
	return &DBError{NotSupported, "unimplemented"}
}

func preCheckKeyValue(key, value *string) error {
	if key != nil && *key == "" {
		return &DBError{InvalidArgument, "key must not be empty"}
	}
	return nil
}

func (db *kvDB) Put(key, value string, options WriteOptions) error {
	if err := preCheckKeyValue(&key, &value); err != nil {
		return err
	}

	// TODO
	return &DBError{NotSupported, "unimplemented"}
}

func (db *kvDB) ReadModifyWrite(key, value string, options WriteOptions) (string, error) {
	if err := preCheckKeyValue(&key, &value); err != nil {
		return "", err
	}

	// TODO
	return "", &DBError{NotSupported, "unimplemented"}
}

func (db *kvDB) Get(key string, options ReadOptions) (string, error) {
	if err := preCheckKeyValue(&key, nil); err != nil {
		return "", err
	}

	// TODO
	return "", &DBError{NotSupported, "unimplemented"}
}

func (db *kvDB) Delete(key string, options WriteOptions) error {
	if err := preCheckKeyValue(&key, nil); err != nil {
		return err
	}

	// TODO
	return &DBError{NotSupported, "unimplemented"}
}
