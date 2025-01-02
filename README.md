# OrderedMap Package

A high-performance, thread-safe implementation of an ordered map data structure in Go, using a doubly linked list for order maintenance and a hash map for fast lookups.

## Features

### Core Functionality
- Thread-safe operations with RWMutex
- O(1) lookups using hash map
- O(1) insertions and deletions using doubly linked list
- Memory-efficient implementation with proper GC handling
- Generic key-value storage using `any` type

### Operations
- `Set`: Add or update key-value pairs - O(1)
- `Get`: Retrieve values by key - O(1)
- `Delete`: Remove key-value pairs - O(1)
- `Clear`: Remove all elements - O(n)
- `Copy`: Create a deep copy - O(n)
- `Has`: Check key existence - O(1)
- `Len`: Get number of elements - O(1)

### Order-Aware Operations
- `Keys`: Get all keys in insertion order - O(n)
- `Values`: Get all values in insertion order - O(n)
- `Range`: Iterate over pairs in order - O(n)
- `String`: Get ordered string representation - O(n)


> **Note:** This OrderedMap implementation is part of a larger data structures project. However, this repo is more comprehensive. For a more comprehensive collection of data structures and algorithms in Go, visit the main repository at [@mstgnz/data-structures](https://github.com/mstgnz/data-structures).


## Usage Examples

### Basic Operations
```go
// Create a new ordered map
om := NewOrderedMap()

// Add key-value pairs
om.Set("first", 1)
om.Set("second", 2)
om.Set("third", 3)

// Get value by key
value, exists := om.Get("second")
if exists {
    fmt.Println(value) // Outputs: 2
}

// Delete a key-value pair
om.Delete("second")

// Check if key exists
exists = om.Has("first") // returns true
```

### Iteration and Order
```go
// Get all keys in order
keys := om.Keys() // ["first", "third"]

// Get all values in order
values := om.Values() // [1, 3]

// Iterate over pairs in order
om.Range(func(key, value any) bool {
    fmt.Printf("%v: %v\n", key, value)
    return true // continue iteration
})
```

## Implementation Details

### Data Structure
- Doubly linked list for order maintenance
- Hash map for O(1) lookups
- Thread-safe with RWMutex

```go
type Node struct {
    Key   any
    Value any
    prev  *Node
    next  *Node
}

type OrderedMap struct {
    mu      sync.RWMutex
    head    *Node
    tail    *Node
    nodeMap map[any]*Node
    length  int
}
```

### Performance Characteristics
- Memory efficient: No slice reallocations
- O(1) operations for basic functions
- Optimized for large datasets
- Efficient garbage collection
- No memory leaks

### Thread Safety
- Read operations can occur concurrently
- Write operations are serialized
- Safe for concurrent access
- Deadlock prevention with deferred unlocks

## Benchmarks
Run benchmarks using:
```bash
go test -bench=. -benchmem
```

### Benchmark Results
Below are the benchmark results on Apple M1 CPU:

```
BenchmarkSet          3384583    305.2 ns/op    141 B/op    3 allocs/op
BenchmarkGet         73489630     15.2 ns/op      0 B/op    0 allocs/op
BenchmarkDelete       8095078    181.1 ns/op      0 B/op    0 allocs/op
BenchmarkRange         706783   1424.0 ns/op      0 B/op    0 allocs/op
BenchmarkCopy          10000  144245.0 ns/op  164603 B/op  1014 allocs/op
```

#### Parallel Operations
```
BenchmarkParallelSet  5840424    218.2 ns/op     33 B/op    2 allocs/op
BenchmarkParallelGet 24743820     82.2 ns/op      0 B/op    0 allocs/op
```

#### Performance Analysis
- **Get Operations**: Extremely fast with ~15ns per operation and zero allocations
- **Set Operations**: Efficient with ~305ns per operation and minimal memory usage
- **Delete Operations**: Quick with ~181ns per operation and no additional memory allocation
- **Range Operations**: Linear time complexity as expected, processing 1000 items in ~1.4µs
- **Copy Operations**: Most resource-intensive at ~144µs per operation with significant memory allocation
- **Parallel Operations**: Shows good scalability with reduced latency under concurrent load

## Testing
Comprehensive test suite including:
- Basic operations
- Edge cases
- Concurrent operations
- Data consistency
- Memory management
- Large dataset handling

Test coverage: 100% of statements

Run tests:
```bash
go test ./...
```

Run tests with coverage:
```bash
go test -cover
```

## Contributing
This project is open-source, and contributions are welcome. Feel free to contribute or provide feedback of any kind.

## License
MIT License - see LICENSE file for details 