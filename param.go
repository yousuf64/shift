package shift

type Param struct {
	key   string
	value string
}

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

func (p *Params) setKeys(keys *[]string) {
	p.keys = keys
	p.values = p.values[:len(*keys)]
}

func (p *Params) appendValue(value string) {
	if p.i >= p.max {
		return
	}
	p.values[p.i] = value
	p.i++
}

func (p *Params) reset() {
	p.i = 0
	p.keys = nil
	p.values = p.values[:0]
}

func (p *Params) Get(k string) string {
	if p.keys != nil {
		for i, key := range *p.keys {
			if key == k {
				return p.values[i]
			}
		}
	}
	return ""
}

func (p *Params) ForEach(f func(k, v string) bool) {
	if p.keys != nil {
		for i := len(*p.keys) - 1; i >= 0; i-- {
			f((*p.keys)[i], p.values[i])
		}
	}
}

func (p *Params) Map() map[string]string {
	params := make(map[string]string, len(*p.keys))

	for i := len(*p.keys) - 1; i >= 0; i-- {
		params[(*p.keys)[i]] = p.values[i]
	}

	return params
}

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

func (p *Params) Copy() *Params {
	cp := new(Params)
	*cp = *p
	return cp
}

// emptyParams is a Params object with 0 capacity, therefore its basically immutable and concurrent safe.
var emptyParams = newParams(0)
