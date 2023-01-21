package dune

import (
	"context"
	"net/http"
)

type Param struct {
	key   string
	value string
}

type Params struct {
	params []Param
	i      int
	max    int
}

func newParams(cap int) *Params {
	return &Params{
		max:    cap,
		params: make([]Param, 0, cap),
	}
}

func (p *Params) set(k, v string) {
	if p.i >= p.max {
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
	for i := len(p.params) - 1; i >= 0; i-- {
		f(p.params[i].key, p.params[i].value)
	}
}

var paramKey = &struct{}{}

func withParamsCtx(ctx context.Context, ps *Params) context.Context {
	return context.WithValue(ctx, paramKey, ps)
}

func hasParamsCtx(ctx context.Context) bool {
	return ctx.Value(paramKey) != nil
}

func paramsFromCtx(ctx context.Context) (*Params, bool) {
	ps, ok := ctx.Value(paramKey).(*Params)
	return ps, ok
}

// emptyParams is a Params object with 0 capacity, therefore its basically immutable and concurrent safe.
var emptyParams = newParams(0)

// Vars returns the http.Request's Params from which route variables can be retrieved.
func Vars(r *http.Request) *Params {
	if ps, ok := paramsFromCtx(r.Context()); ok {
		return ps
	}

	return emptyParams
}
