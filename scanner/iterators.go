package scanner

import (
	"reflect"

	reflectx "github.com/go-courier/x/reflect"
)

type ScanIterator interface {
	// New a ptr value for scan
	New() interface{}
	// Next For receive scanned value
	Next(v interface{}) error
}

func ScanIteratorFor(v interface{}) (ScanIterator, error) {
	switch x := v.(type) {
	case ScanIterator:
		return x, nil
	default:
		tpe := reflectx.Deref(reflect.TypeOf(v))

		if tpe.Kind() == reflect.Slice && tpe.Elem().Kind() != reflect.Uint8 {
			return &SliceScanIterator{
				elemType: tpe.Elem(),
				rv:       reflectx.Indirect(reflect.ValueOf(v)),
			}, nil
		}

		return &SingleScanIterator{target: v}, nil
	}
}

type SliceScanIterator struct {
	elemType reflect.Type
	rv       reflect.Value
}

func (s *SliceScanIterator) New() interface{} {
	return reflectx.New(s.elemType).Addr().Interface()
}

func (s *SliceScanIterator) Next(v interface{}) error {
	s.rv.Set(reflect.Append(s.rv, reflect.ValueOf(v).Elem()))
	return nil
}

type SingleScanIterator struct {
	target     interface{}
	hasResults bool
}

func (s *SingleScanIterator) New() interface{} {
	return s.target
}

func (s *SingleScanIterator) Next(v interface{}) error {
	s.hasResults = true
	return nil
}

func (s *SingleScanIterator) MustHasRecord() bool {
	return s.hasResults
}
