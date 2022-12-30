package dune

type Param struct {
	key   string
	value string
}

type Params struct {
	params []Param
	i      uint8
	max    int
}

func newParams(cap int) *Params {
	return &Params{
		max:    cap,
		params: make([]Param, 0, cap),
	}
}

func (p *Params) set(k, v string) {
	if int(p.i) >= p.max {
		return
	}

	p.params = p.params[:p.i+1]
	p.params[p.i] = Param{k, v}
	p.i++
}

func (p *Params) reset() {
	p.params = p.params[:0]
	p.i = 0
}

func (p *Params) Get(k string) string {
	for _, kv := range p.params {
		if kv.key == k {
			return kv.value
		}
	}
	return ""
}

func (p *Params) ForEach(f func(k, v string) bool) {
	for _, kv := range p.params {
		f(kv.key, kv.value)
	}
}
