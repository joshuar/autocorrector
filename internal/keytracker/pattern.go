package keytracker

import "bytes"

type patternBuf struct {
	buf      *bytes.Buffer
	len, idx int
}

func newPatternBuf(size int) *patternBuf {
	return &patternBuf{
		buf: new(bytes.Buffer),
		len: size,
		idx: 0,
	}
}

func (pb *patternBuf) write(r rune) {
	if pb.idx == 3 {
		pb.buf.Reset()
		pb.idx = 0
	}
	pb.buf.WriteRune(r)
	pb.idx++
}

func (pb *patternBuf) match(s string) bool {
	return bytes.HasSuffix(pb.buf.Bytes(), []byte(s))
}
