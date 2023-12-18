package shift

import (
	"fmt"
	"testing"
)

func TestParams_Get(t *testing.T) {
	ip := newInternalParams(3)
	ip.setKeys(&[]string{"k3", "k2", "k1"})
	ip.appendValue("xyz")
	ip.appendValue("bar")
	ip.appendValue("foo")

	p := newParams(ip)

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
	ip := newInternalParams(3)
	ip.setKeys(&[]string{"k3", "k2", "k1"})
	ip.appendValue("xyz")
	ip.appendValue("bar")
	ip.appendValue("foo")

	p := newParams(ip)

	tests := []struct{ k, v string }{
		{"k1", "foo"},
		{"k2", "bar"},
		{"k3", "xyz"},
	}

	i := 0
	p.ForEach(func(k, v string) {
		assert(t, k == tests[i].k, fmt.Sprintf("key > expected: %s, got: %s", tests[i].k, k))
		assert(t, v == tests[i].v, fmt.Sprintf("value > expected: %s, got: %s", tests[i].v, v))
		i++
	})
	assert(t, i == 3, fmt.Sprintf("i > expected: %d, got: %d", 3, i))
}

func TestParams_Map(t *testing.T) {
	ip := newInternalParams(3)
	ip.setKeys(&[]string{"k3", "k2", "k1"})
	ip.appendValue("xyz")
	ip.appendValue("bar")
	ip.appendValue("foo")

	p := newParams(ip)

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
	assert(t, len(params) == 3, fmt.Sprintf("map length > expected: %d, got: %d", 3, len(params)))
}

func TestParams_Slice(t *testing.T) {
	ip := newInternalParams(3)
	ip.setKeys(&[]string{"k3", "k2", "k1"})
	ip.appendValue("xyz")
	ip.appendValue("bar")
	ip.appendValue("foo")

	p := newParams(ip)

	tests := []struct{ k, v string }{
		{"k1", "foo"},
		{"k2", "bar"},
		{"k3", "xyz"},
	}

	params := p.Slice()
	for i := 0; i < len(tests); i++ {
		assert(t, params[i].Key == tests[i].k, fmt.Sprintf("key at index %d > expected: %s, got: %s", i, tests[i].k, params[i].Key))
		assert(t, params[i].Value == tests[i].v, fmt.Sprintf("value at index %d > expected: %s, got: %s", i, tests[i].v, params[i].Value))
	}
	assert(t, len(params) == 3, fmt.Sprintf("params count > expected: %d, got: %d", 3, len(params)))
}

func TestParams_Copy(t *testing.T) {
	t.Run("copied Params should not be equal to source Params", func(t *testing.T) {
		ip := newInternalParams(1)
		ip.setKeys(&[]string{"foo"})
		ip.appendValue("bar")

		p := newParams(ip)
		cp := p.Copy()

		assert(t, cp != p, "expected to be unequal")
		assert(t, cp.internal != p.internal, "expected internal to be unequal")
	})

	t.Run("copied Params should not be affected when source is reset", func(t *testing.T) {
		ip := newInternalParams(1)
		ip.setKeys(&[]string{"foo"})
		ip.appendValue("bar")

		p := newParams(ip)
		cp := p.Copy()
		ip.reset() // resets internal of source

		val := cp.Get("foo")
		assert(t, val == "bar", fmt.Sprintf("expected: bar, got: %s", val))
	})

	t.Run("copied Params should not be affected when source is reset and reused", func(t *testing.T) {
		ip := newInternalParams(2)
		ip.setKeys(&[]string{"foo", "woo"})
		ip.appendValue("bar")
		ip.appendValue("abc")

		p := newParams(ip)
		cp := p.Copy()
		ip.reset() // resets internal of source
		ip.setKeys(&[]string{"foo"})
		ip.appendValue("xyz")

		val := cp.Get("foo")
		assert(t, val == "bar", fmt.Sprintf("expected: bar, got: %s", val))
	})
}

func BenchmarkParams_Copy(b *testing.B) {
	b.Run("with non-<nil> internal", func(b *testing.B) {
		ip := newInternalParams(10)
		ip.setKeys(&[]string{"1", "2", "3"})
		ip.appendValue("abc")
		ip.appendValue("xyz")
		ip.appendValue("cap")

		p := newParams(ip)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			p.Copy()
		}
	})

	b.Run("with <nil> internal", func(b *testing.B) {
		p := newParams(nil)

		b.ReportAllocs()
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			p.Copy()
		}
	})
}

func BenchmarkInternalParams_DeepCopy(b *testing.B) {
	ip := newInternalParams(10)
	ip.setKeys(&[]string{"1", "2", "3"})
	ip.appendValue("abc")
	ip.appendValue("xyz")
	ip.appendValue("cap")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ip.deepCopy()
	}
}
