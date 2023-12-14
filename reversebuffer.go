package shift

type reverseBuffer interface {
	WriteString(s string)
	String() string
}

type sizedReverseBuffer struct {
	offset int
	b      []byte
}

func newSizedReverseBuffer(size int) *sizedReverseBuffer {
	return &sizedReverseBuffer{
		offset: size,
		b:      make([]byte, size),
	}
}

func (buf *sizedReverseBuffer) WriteString(s string) {
	for i, j := buf.offset-1, len(s)-1; i >= 0 && j >= 0; i, j = i-1, j-1 {
		buf.b[i] = s[j]
	}
	buf.offset -= len(s)
	if buf.offset < 0 {
		buf.offset = 0
	}
}

func (buf *sizedReverseBuffer) String() string {
	return bytesToString(buf.b[buf.offset:])
}

type reverseBuffer128 struct {
	offset int
	b      [128]byte
}

func newReverseBuffer128() *reverseBuffer128 {
	return &reverseBuffer128{
		offset: 128,
		b:      [128]byte{},
	}
}

func (buf *reverseBuffer128) WriteString(s string) {
	for i, j := buf.offset-1, len(s)-1; i >= 0 && j >= 0; i, j = i-1, j-1 {
		buf.b[i] = s[j]
	}
	buf.offset -= len(s)
	if buf.offset < 0 {
		buf.offset = 0
	}
}

func (buf *reverseBuffer128) String() string {
	return bytesToString(buf.b[buf.offset:])
}
