package shift

import "sync"

// Param is a key-value pair of request's route params.
type Param struct {
	Key   string
	Value string
}

// Params stores the request's route params.
//
// When passing [Params] to a goroutine or if it's intended to use [Params] beyond the request lifecycle,
// make sure to pass/store a copy of [Params] using [Params.Copy].
//
// While [Params] is concurrent safe, it is not designed to be reliably used beyond the request lifecycle.
// The reason being underlying store of [Params] is pooled into a [sync.Pool] when the request is completed,
// which can potentially be used by another request.
type Params struct {
	internal *internalParams
}

func newParams(internalParams *internalParams) Params {
	return Params{
		internal: internalParams,
	}
}

// Get retrieves the value associated with the provided key.
func (p *Params) Get(key string) string {
	if p.internal == nil {
		return ""
	}
	return p.internal.get(key)
}

// ForEach iterates through Params in the order params are defined in the route.
func (p *Params) ForEach(fn func(k, v string)) {
	if p.internal == nil {
		return
	}
	p.internal.forEach(fn)
}

// Map returns Params mapped into a [key]value map.
func (p *Params) Map() map[string]string {
	if p.internal == nil {
		return nil
	}
	return p.internal.kvMap()
}

// Slice returns a slice of Param in the order params are defined in the route.
func (p *Params) Slice() []Param {
	if p.internal == nil {
		return nil
	}
	return p.internal.slice()
}

// Len returns the length of Params.
func (p *Params) Len() int {
	if p.internal == nil {
		return 0
	}
	return len(p.internal.values)
}

// Copy returns a deep-copy of Params.
func (p *Params) Copy() Params {
	if p.internal == nil {
		return *p
	}
	return Params{
		internal: p.internal.deepCopy(),
	}
}

func (p *Params) release(pool *sync.Pool) {
	p.internal.reset()
	pool.Put(p.internal)
	p.internal = nil
}

// internalParams is the underlying store of [Params]. To reduce allocations, internalParams are pooled into a [sync.Pool].
type internalParams struct {
	i      int
	max    int       // Is the capacity of values. It's meant to prevent overflows.
	keys   *[]string // Value of keys is immutable (created once at startup and passed around). Therefore, it can be shared by different internalParams concurrently.
	values []string
}

func newInternalParams(cap int) *internalParams {
	return &internalParams{
		i:      0,
		max:    cap,
		keys:   nil,
		values: make([]string, 0, cap),
	}
}

// setKeys replaces keys with the provided keys and expands/shrinks values to the keys' length.
func (p *internalParams) setKeys(keys *[]string) {
	p.keys = keys
	p.values = p.values[:len(*keys)]
}

// appendValue appends a value if the max capacity is not reached and increases the counter.
// It accepts values irrespective of the keys' length.
func (p *internalParams) appendValue(value string) {
	if p.i >= p.max {
		return
	}
	p.values[p.i] = value
	p.i++
}

// reset resets the state.
func (p *internalParams) reset() {
	p.i = 0
	p.keys = nil
	p.values = p.values[:0]
}

// get retrieves the value associated with the provided key.
func (p *internalParams) get(key string) string {
	if p.keys != nil {
		for i, k := range *p.keys {
			if k == key {
				return p.values[i]
			}
		}
	}
	return ""
}

// forEach iterates through internalParams in the order params are defined in the route.
func (p *internalParams) forEach(fn func(k, v string)) {
	if p.keys != nil {
		for i := len(*p.keys) - 1; i >= 0; i-- {
			fn((*p.keys)[i], p.values[i])
		}
	}
}

// kvMap returns internalParams mapped into a [key]value map.
func (p *internalParams) kvMap() map[string]string {
	params := make(map[string]string, len(*p.keys))

	for i := len(*p.keys) - 1; i >= 0; i-- {
		params[(*p.keys)[i]] = p.values[i]
	}

	return params
}

// slice returns a slice of Param in the order params are defined in the route.
func (p *internalParams) slice() []Param {
	params := make([]Param, 0, len(*p.keys))

	for i := len(*p.keys) - 1; i >= 0; i-- {
		params = append(params, Param{
			Key:   (*p.keys)[i],
			Value: p.values[i],
		})
	}

	return params
}

// deepCopy returns a deep-copy of internalParams.
func (p *internalParams) deepCopy() *internalParams {
	values := make([]string, len(p.values))
	copy(values, p.values)

	return &internalParams{
		i:      p.i,
		max:    p.max,
		keys:   p.keys,
		values: values,
	}
}
