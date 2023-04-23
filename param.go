package shift

// Param is a key-value pair of request's route params.
type Param struct {
	key   string
	value string
}

// Params stores the request's route params.
//
// When passing Params to a goroutine, make to sure pass a copy (use Copy method)
// instead of the original Params object. The reason being Params is pooled into a sync.Pool when the
// request is completed.
type Params struct {
	i      int
	max    int
	keys   *[]string
	values []string
}

func newParams(cap int) *Params {
	return &Params{
		i:      0,
		max:    cap,
		keys:   nil,
		values: make([]string, 0, cap),
	}
}

// setKeys replaces keys with the provided keys and expands/shrinks values to the keys' length.
func (p *Params) setKeys(keys *[]string) {
	p.keys = keys
	p.values = p.values[:len(*keys)]
}

// appendValue appends a value if the max capacity is not reached and increases the counter.
// It accepts values irrespective of the keys' length.
func (p *Params) appendValue(value string) {
	if p.i >= p.max {
		return
	}
	p.values[p.i] = value
	p.i++
}

// reset resets the state.
func (p *Params) reset() {
	p.i = 0
	p.keys = nil
	p.values = p.values[:0]
}

// Get retrieves the value associated with the provided key.
func (p *Params) Get(key string) string {
	if p.keys != nil {
		for i, k := range *p.keys {
			if k == key {
				return p.values[i]
			}
		}
	}
	return ""
}

// ForEach iterates through Params in the order params are defined in the route.
func (p *Params) ForEach(fn func(k, v string)) {
	if p.keys != nil {
		for i := len(*p.keys) - 1; i >= 0; i-- {
			fn((*p.keys)[i], p.values[i])
		}
	}
}

// Map returns Params mapped into a [key]value map.
func (p *Params) Map() map[string]string {
	params := make(map[string]string, len(*p.keys))

	for i := len(*p.keys) - 1; i >= 0; i-- {
		params[(*p.keys)[i]] = p.values[i]
	}

	return params
}

// Slice returns a slice of Param in the order params are defined in the route.
func (p *Params) Slice() []Param {
	params := make([]Param, 0, len(*p.keys))

	for i := len(*p.keys) - 1; i >= 0; i-- {
		params = append(params, Param{
			key:   (*p.keys)[i],
			value: p.values[i],
		})
	}

	return params
}

// Copy returns a copy of Params.
func (p *Params) Copy() *Params {
	values := make([]string, len(p.values))
	copy(values, p.values)

	return &Params{
		i:      p.i,
		max:    p.max,
		keys:   p.keys,
		values: values,
	}
}

// emptyParams is a Params object with 0 capacity.
// This should be passed to the HandlerFunc (within Route arg) of static routes to ensure Route.*Params
// is always non-nil. The same instance can be passed to any number of HandlerFunc simultaneously since the Params is
// immutable through the public API, hence concurrent safe.
//
// Also, avoid pooling emptyParams as it cannot take writes.
var emptyParams = newParams(0)
