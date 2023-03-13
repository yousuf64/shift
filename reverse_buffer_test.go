package shift

import (
	"fmt"
	"strings"
	"testing"
)

func TestReverseBuffer128(t *testing.T) {
	paths := map[string][]string{
		"/xxx/yyy/zzz":         {"/zzz", "/yyy", "/xxx"},
		"/abc/xyz/ooo/mmm":     {"/mmm", "/ooo", "/xyz", "/abc"},
		"/aaa/bbb/ccc/ddd/eee": {"/eee", "/ddd", "/ccc", "/bbb", "/aaa"},
	}

	for path, parts := range paths {
		buf := newReverseBuffer128()

		for _, part := range parts {
			buf.WriteString(part)
		}

		str := buf.String()
		assert(t, str == path, fmt.Sprintf("expected: %s, got: %s", path, str))
	}
}

func TestReverseBuffer128_PreventOverflow(t *testing.T) {
	buf := newReverseBuffer128()
	var sb strings.Builder

	for i := 0; i < 125; i++ {
		sb.WriteRune('a')
	}

	buf.WriteString(sb.String())
	buf.WriteString("qwexyz")
	buf.WriteString("foo")

	expectation1 := 128
	expectation2 := "xyz" + sb.String()

	str := buf.String()
	length := len(str)
	assert(t, length == expectation1, fmt.Sprintf("length expected: %d, got: %d", expectation1, length))
	assert(t, str == expectation2, fmt.Sprintf("string expected: %s, got: %s", expectation2, str))
}

func TestSizedReverseBuffer(t *testing.T) {
	paths := map[string][]string{
		"/xxx/yyy/zzz":         {"/zzz", "/yyy", "/xxx"},
		"/abc/xyz/ooo/mmm":     {"/mmm", "/ooo", "/xyz", "/abc"},
		"/aaa/bbb/ccc/ddd/eee": {"/eee", "/ddd", "/ccc", "/bbb", "/aaa"},
	}

	for path, parts := range paths {
		buf := newSizedReverseBuffer(len(path))

		for _, part := range parts {
			buf.WriteString(part)
		}

		str := buf.String()
		assert(t, str == path, fmt.Sprintf("expected: %s, got: %s", path, str))
	}
}

func TestSizedReverseBuffer_PreventOverflow(t *testing.T) {
	expectation1 := 5
	expectation2 := "yzabc"
	buf := newSizedReverseBuffer(expectation1)

	buf.WriteString("abc")
	buf.WriteString("xyz")
	buf.WriteString("jkl")

	str := buf.String()
	length := len(str)
	assert(t, length == expectation1, fmt.Sprintf("length expected: %d, got: %d", expectation1, length))
	assert(t, str == expectation2, fmt.Sprintf("string expected: %s, got: %s", expectation2, str))
}
