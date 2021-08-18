package ql

import "fmt"

type QLString interface {
	Val() string
	Values() []string
}
type QLBool interface {
	Val() bool
}
type QLInt interface {
	Val() int
}
type QLInt32 interface {
	Val() int32
}
type QLInt64 interface {
	Val() int64
}
type QLFloat64 interface {
	Val() float64
}

type stringT struct {
	string  *string
	strings []*string
}

func (s *stringT) Val() string {
	val := ""
	if s.string != nil {
		val = *s.string
	}

	return fmt.Sprintf("%s", val)
}

func (s *stringT) Values() []string {
	return nil // TODO:
}

func String(val ...string) QLString {
	return &stringT{string: &val[0]} // TODO:
}

func StringPtr(val ...*string) QLString {
	return &stringT{string: val[0]} // TODO:
}

type boolT struct {
	bool *bool
}

func (b *boolT) Val() bool {
	if b.bool == nil {
		return false
	}

	return *b.bool
}

func Bool(val bool) QLBool {
	return &boolT{bool: &val}
}

func BoolPtr(val *bool) QLBool {
	return &boolT{bool: val}
}

type intT struct {
	int *int
}

func (i *intT) Val() int {
	if i.int == nil {
		return 0
	}

	return *i.int
}

func Int(val int) QLInt {
	return &intT{int: &val}
}

func IntPtr(val *int) QLInt {
	return &intT{int: val}
}

type int32T struct {
	int32 *int32
}

func (i *int32T) Val() int32 {
	if i.int32 == nil {
		return 0
	}

	return *i.int32
}

func Int32(val int32) QLInt32 {
	return &int32T{int32: &val}
}

func Int32Ptr(val *int32) QLInt32 {
	return &int32T{int32: val}
}

type int64T struct {
	int64 *int64
}

func (i *int64T) Val() int64 {
	if i.int64 == nil {
		return 0
	}

	return *i.int64
}

func Int64(val int64) QLInt64 {
	return &int64T{int64: &val}
}

func Int64Ptr(val *int64) QLInt64 {
	return &int64T{int64: val}
}

type float64T struct {
	float64 *float64
}

func (i *float64T) Val() float64 {
	if i.float64 == nil {
		return 0
	}

	return *i.float64
}

func Float64(val float64) QLFloat64 {
	return &float64T{float64: &val}
}

func Float64Ptr(val *float64) QLFloat64 {
	return &float64T{float64: val}
}
