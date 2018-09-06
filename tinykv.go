/*
 * Copyright 2018 The TinyKV Authors.
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

package tinykv

import "fmt"

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
type ErrorCode int

const (
	NotFound        ErrorCode = 0x0
	Corruption      ErrorCode = 0x1
	NotSupported    ErrorCode = 0x2
	InvalidArgument ErrorCode = 0x3
	IOError         ErrorCode = 0x4
)

// DBError is error type used in TinyKV project.\
type DBError struct {
	Code   ErrorCode
	Status string
}

func (err *DBError) Error() string {
	return fmt.Sprintf("error code: %d, details: %s", err.Code, err.Status)
}

func IsNotFoundError(err error) bool {
	if dbErr, ok := err.(*DBError); ok {
		return dbErr.Code == NotFound
	}
	return false
}

func IsCorruptionError(err error) bool {
	if dbErr, ok := err.(*DBError); ok {
		return dbErr.Code == Corruption
	}
	return false
}

func IsNotSupportedError(err error) bool {
	if dbErr, ok := err.(*DBError); ok {
		return dbErr.Code == NotSupported
	}
	return false
}

func IsInvalidArgumentError(err error) bool {
	if dbErr, ok := err.(*DBError); ok {
		return dbErr.Code == InvalidArgument
	}
	return false
}

func IsIOError(err error) bool {
	if dbErr, ok := err.(*DBError); ok {
		return dbErr.Code == IOError
	}
	return false
}

/*******************************************************************************
  Here we define Options, WriteOptions and ReadOptions to be used in DB.
 ******************************************************************************/
type CompressionType int

const (
	NoCompression     CompressionType = 0x0
	SnappyCompression CompressionType = 0x1
)

type Options struct {
	CreateIfMissing bool
	ErrorIfExists   bool
	ParanoidChecks  bool
	WriteBufferSize int64
	MaxOpenFiles    int
	Compression     CompressionType
}

type ReadOptions struct {
	VerifyChecksum bool
	FillCache      bool
}

type WriteOptions struct {
	Sync bool
}

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

func NewReadOptions() ReadOptions {
	return ReadOptions{
		VerifyChecksum: false,
		FillCache:      false,
	}
}

func NewWriteOptions() WriteOptions {
	return WriteOptions{
		Sync: false,
	}
}

/*******************************************************************************
  Here we define the DB interface and kvDB struct which implements DB.
 ******************************************************************************/

type DB interface {
	// Open the database. Returns nil on success, and an error otherwise.
	Open() error

	// Close the database. Returns nil on success, and an error otherwise.
	// Operations on a closed database will get a not supported error. You can distinguish it
	// using IsNotSupportedError(error).
	Close() error

	// Put key-value pair into the database. Returns nil on success, and an error otherwise.
	// If key or value is nil, it returns a invalid argument error. You can distinguish it using
	// IsInvalidArgumentError(error).
	Put(key, value string, options WriteOptions) error

	// Perform a read-modify-write atomic operation on database entry for key.
	// Returns old value and nil on success, and nil and an error otherwise.
	// If there's no such key, the old value returned should be nil.
	// If key or value is nil, it returns a invalid argument error (and can be distinguished by
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

type kvDB struct {
	// force implements DB interface
	DB

	// database name
	name string
	// database options
	options Options
}

func NewDB(name string, options Options) DB {
	return kvDB{options: options, name: name}
}

func DestoryDB(name string, options Options) error {
	// TODO
	return &DBError{NotSupported, "unimplemented"}
}

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

func (db *kvDB) Put(key, value string, options WriteOptions) error {
	// TODO
	return &DBError{NotSupported, "unimplemented"}
}

func (db *kvDB) ReadModifyWrite(key, value string, options WriteOptions) error {
	// TODO
	return &DBError{NotSupported, "unimplemented"}
}

func (db *kvDB) Get(key string, options ReadOptions) (string, error) {
	// TODO
	return "", &DBError{NotSupported, "unimplemented"}
}

func (db *kvDB) Delete(key string, options WriteOptions) error {
	// TODO
	return &DBError{NotSupported, "unimplemented"}
}
