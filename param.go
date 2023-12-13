package shift

// Param is a key-value pair of request's route params.
type Param struct {
	Key   string
	Value string
}

type Params struct {
	internal *internalParams
}

// Get retrieves the value associated with the provided key.
func (p Params) Get(key string) string {
	return p.internal.Get(key)
}

// ForEach iterates through internalParams in the order params are defined in the route.
func (p Params) ForEach(fn func(k, v string)) {
	p.internal.ForEach(fn)
}

// Map returns internalParams mapped into a [key]value map.
func (p Params) Map() map[string]string {
	return p.internal.Map()
}

// Slice returns a slice of Param in the order params are defined in the route.
func (p Params) Slice() []Param {
	return p.internal.Slice()
}

// Copy returns a copy of internalParams.
func (p Params) Copy() Params {
	values := make([]string, len(p.internal.values))
	copy(values, p.internal.values)

	return Params{&internalParams{
		i:      p.internal.i,
		max:    p.internal.max,
		keys:   p.internal.keys,
		values: values,
	}}
}

// internalParams stores the request's route params.
//
// When passing internalParams to a goroutine, make to sure pass a copy (use Copy method)
// instead of the original internalParams object. The reason being internalParams is pooled into a sync.Pool when the
// request is completed.
type internalParams struct {
	i      int
	max    int
	keys   *[]string // Immutable (created once at startup and passed around).
	values []string
}

func newParams(cap int) *internalParams {
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

// Get retrieves the value associated with the provided key.
func (p *internalParams) Get(key string) string {
	if p.keys != nil {
		for i, k := range *p.keys {
			if k == key {
				return p.values[i]
			}
		}
	}
	return ""
}

// ForEach iterates through internalParams in the order params are defined in the route.
func (p *internalParams) ForEach(fn func(k, v string)) {
	if p.keys != nil {
		for i := len(*p.keys) - 1; i >= 0; i-- {
			fn((*p.keys)[i], p.values[i])
		}
	}
}

// Map returns internalParams mapped into a [key]value map.
func (p *internalParams) Map() map[string]string {
	params := make(map[string]string, len(*p.keys))

	for i := len(*p.keys) - 1; i >= 0; i-- {
		params[(*p.keys)[i]] = p.values[i]
	}

	return params
}

// Slice returns a slice of Param in the order params are defined in the route.
func (p *internalParams) Slice() []Param {
	params := make([]Param, 0, len(*p.keys))

	for i := len(*p.keys) - 1; i >= 0; i-- {
		params = append(params, Param{
			Key:   (*p.keys)[i],
			Value: p.values[i],
		})
	}

	return params
}

// Copy returns a copy of internalParams.
func (p *internalParams) Copy() *internalParams {
	values := make([]string, len(p.values))
	copy(values, p.values)

	return &internalParams{
		i:      p.i,
		max:    p.max,
		keys:   p.keys,
		values: values,
	}
}

// emptyParams is a internalParams object with 0 capacity.
// This should be passed to the HandlerFunc (within Route arg) of static routes to ensure Route.*internalParams
// is always non-nil. The same instance can be passed to any number of HandlerFunc simultaneously since the internalParams is
// immutable through the public API, hence concurrent safe.
//
// Also, avoid pooling emptyParams as it cannot take writes.
var emptyParams = newParams(0)

func init() {
	emptyParams.setKeys(&[]string{}) // ensures internalParams.keys is non-nil. Prevents from panicking on internalParams.Slice() and internalParams.Map().
}
