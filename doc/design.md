# Features and Layers

We want to make TinyKV a **concurrent-safe**, **thread-efficient** key-value store and are going to implement the
following interface:

```go
type DB interface {
	// Open the database. Returns nil on success., and an error otherwise.
	Open() error

	// Close the database. Returns nil on success, and an error otherwise.
	// Operations on a closed database will get a not supported error. You can distinguish it
	// using IsNotSupportedError(error).
	Close() error

	// Put key-value pair into the database. Returns nil on success, and an error otherwise.
	// If key or value is nil, it returns a invalid argument error. You can distinguish it using
	// IsInvalidArgumentError(error).
	Put(key, value string, options WriteOptions) error

	// Get value of specified key from the database. Returns value and nil on success,
	// and nil and an error otherwise.
	// If key is not found, you will get an not found error. You can distinguish it
	// using IsNotFoundError(error).
	Get(key string, options ReadOptions) (string, error)

	// Delete the database entry for key. Returns nil on success, and an error otherwise.
	// It is not an error if key does not exists in the database.
	Delete(key string, options WriteOptions) error
}
```

Also, we want to optimize for **write** and **sequential read**, and provide support for quick read
for cascading keys like "/university/nju/course/algorithms/students", so we decide to use LSM-Tree as our on-disk data structure.

Furthermore, we think we should have convenient interfaces along with the programming interface. So we decides
to provide two interfaces: a REPL commandline interface and a RPC remote interface.

Since TinyKV should be concurrent-safe, we can manipulate entries concurrently from different processes.

In summary, the major features are

1. Concurrent safe and thread efficient key-value store
2. Optimize for write and sequential read
3. Provide a REPL commandline interface
4. Provide a RPC remote interface

Here's an overview of layers involved in TinyKV:

```
+-----------------+--------------+
|   TinyKV REPL   | TinyKV RPC   |
+-----------------+--------------+
|     TinyKV Direct Interface    |
+--------------------------------+
|    Log Structured Merge Tree   |
+--------------------------------+
| File Manipulation Abstractions |
+--------------------------------+
|           File System          |
+--------------------------------+
|               OS               |
+--------------------------------+
```

We define a `File Manipulation Abstractions` layer to achieve a further convenience to port TinyKV to other
platforms like Windows.

# Open/Close A Database

```go
// Before openning the database, create a new DB instance
dbName := "tinykv"
db := tinykv.NewDB(dbName, tinykv.NewOptions())

// Open the database
if err := db.Open(); err != nil {
	// Handle the error and exit
}

// Manipulate the database
// ...

// Close the database
if err := db.Close(); err != nil {
	// Log the error
}
```

# Reads and Writes

```go
// Create and open a database to "db"
key := "key1"
// Put a new value to database entry for key
if err := db.Put(key, "value1"); err != nil {
	// Handle the error
}

// Get the value for key
if value, err := db.Get(key); err == nil {
	// assert value == "value1", if there's no other process modify this entry
	fmt.Println("%s: %s", key, value)
} else {
	// Handle the error
}

// Delete the entry
if err := db.Delete(key); err != nil {
	// Log the error
}
```