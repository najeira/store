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
// Get returns the element.
value := store.Get(key, func() interface{} {
    // called when the element is not stored.
    // returning value will be stored.
    return calcNewValue(key)
})
```

```go
// delete the element.
store.Del(key)
```

## License

MIT
