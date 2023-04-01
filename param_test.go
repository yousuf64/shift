package shift

import (
	"fmt"
	"testing"
)

func TestParams_Get(t *testing.T) {
	p := newParams(3)
	p.setKeys(&[]string{"k3", "k2", "k1"})
	p.appendValue("xyz")
	p.appendValue("bar")
	p.appendValue("foo")

	tests := map[string]string{
		"k3": "xyz",
		"k2": "bar",
		"k1": "foo",
	}

	for k, v := range tests {
		val := p.Get(k)
		assert(t, val == v, fmt.Sprintf("expected: %s, got: %s", v, val))
	}
}

func TestParams_ForEach(t *testing.T) {
	p := newParams(3)
	p.setKeys(&[]string{"k3", "k2", "k1"})
	p.appendValue("xyz")
	p.appendValue("bar")
	p.appendValue("foo")

	tests := []struct{ k, v string }{
		{"k1", "foo"},
		{"k2", "bar"},
		{"k3", "xyz"},
	}

	i := 0
	p.ForEach(func(k, v string) bool {
		assert(t, k == tests[i].k, fmt.Sprintf("key > expected: %s, got: %s", tests[i].k, k))
		assert(t, v == tests[i].v, fmt.Sprintf("value > expected: %s, got: %s", tests[i].v, v))
		i++
		return true
	})
}

func TestParams_Map(t *testing.T) {
	p := newParams(3)
	p.setKeys(&[]string{"k3", "k2", "k1"})
	p.appendValue("xyz")
	p.appendValue("bar")
	p.appendValue("foo")

	tests := map[string]string{
		"k1": "foo",
		"k2": "bar",
		"k3": "xyz",
	}

	params := p.Map()
	for k, v := range tests {
		val := params[k]
		assert(t, val == v, fmt.Sprintf("value for key %s, expected: %s, got: %s", k, v, val))
	}
}

func TestParams_Slice(t *testing.T) {
	p := newParams(3)
	p.setKeys(&[]string{"k3", "k2", "k1"})
	p.appendValue("xyz")
	p.appendValue("bar")
	p.appendValue("foo")

	tests := []struct{ k, v string }{
		{"k1", "foo"},
		{"k2", "bar"},
		{"k3", "xyz"},
	}

	params := p.Slice()
	for i := 0; i < len(tests); i++ {
		assert(t, params[i].key == tests[i].k, fmt.Sprintf("key at index %d > expected: %s, got: %s", i, tests[i].k, params[i].key))
		assert(t, params[i].value == tests[i].v, fmt.Sprintf("value at index %d > expected: %s, got: %s", i, tests[i].v, params[i].value))
	}
}

func TestParams_Copy(t *testing.T) {
	t.Run("Copy should have a different memory address", func(t *testing.T) {
		p := newParams(1)
		p.setKeys(&[]string{"foo"})
		p.appendValue("bar")

		cp := p.Copy()

		assert(t, cp != p, "expected different memory addresses")
	})

	t.Run("Copy should not be affected when source is reset", func(t *testing.T) {
		p := newParams(1)
		p.setKeys(&[]string{"foo"})
		p.appendValue("bar")

		cp := p.Copy()
		p.reset()

		val := cp.Get("foo")
		assert(t, val == "bar", fmt.Sprintf("expected: bar, got: %s", val))
	})

	t.Run("Copy should not be affected when source is reset and reused", func(t *testing.T) {
		p := newParams(2)
		p.setKeys(&[]string{"foo", "woo"})
		p.appendValue("bar")
		p.appendValue("abc")

		cp := p.Copy()
		p.reset()
		p.setKeys(&[]string{"foo"})
		p.appendValue("xyz")

		val := cp.Get("foo")
		assert(t, val == "bar", fmt.Sprintf("expected: bar, got: %s", val))
	})
}

func BenchmarkParams_Copy(b *testing.B) {
	p := newParams(10)
	p.setKeys(&[]string{"1", "2", "3"})
	p.appendValue("abc")
	p.appendValue("xyz")
	p.appendValue("cap")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		p.Copy()
	}
}
