//go:build unit

package service

func testPtrFloat64(value float64) *float64 {
	return &value
}
