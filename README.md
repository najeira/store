# store

## Import

```go
import "github.com/najeira/store"
```

## Usage

```go
// allocate *store.Store on global.
var store = store.New()
```

```go
// Get gets the element.
value, ok := store.Get(key)
```

```go
// Set sets the element.
store.Set(key, value)
```

```go
// Del deletes the element.
store.Del(key)
```

```go
// Fetch returns the element.
value := store.Fetch(key, func() interface{} {
    // called when the element is not stored.
    // returning value will be stored.
    return calcNewValue(key)
})
```

## License

MIT
