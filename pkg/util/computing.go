package util

type Byte float64

func (b Byte) Int64() int64 {
	return int64(b)
}

func (b Byte) Int() int {
	return int(b)
}

const (
	_       = iota
	KB Byte = 1 << (10 * iota)
	MB
	GB
)
