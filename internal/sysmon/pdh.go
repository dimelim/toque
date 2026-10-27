package sysmon

import (
	"errors"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	pdh                      = windows.NewLazySystemDLL("pdh.dll")
	procPdhOpenQuery         = pdh.NewProc("PdhOpenQueryW")
	procPdhAddEnglishCounter = pdh.NewProc("PdhAddEnglishCounterW")
	procPdhCollectQueryData  = pdh.NewProc("PdhCollectQueryData")
	procPdhGetFormattedValue = pdh.NewProc("PdhGetFormattedCounterValue")
	procPdhGetFormattedArray = pdh.NewProc("PdhGetFormattedCounterArrayW")
	procPdhCloseQuery        = pdh.NewProc("PdhCloseQuery")
)

const (
	pdhFmtDouble   = 0x00000200
	pdhFmtNoCap100 = 0x00008000
	pdhMoreData    = 0x800007D2
)

type fmtValue struct {
	CStatus uint32
	_       uint32
	Double  float64
}

type fmtItem struct {
	Name  *uint16
	Value fmtValue
}

type query struct {
	h uintptr
}

func openQuery() (*query, error) {
	var h uintptr
	if r, _, _ := procPdhOpenQuery.Call(0, 0, uintptr(unsafe.Pointer(&h))); r != 0 {
		return nil, errors.New("PdhOpenQuery failed")
	}
	return &query{h: h}, nil
}

func (q *query) add(path string) (uintptr, error) {
	p, _ := windows.UTF16PtrFromString(path)
	var c uintptr
	if r, _, _ := procPdhAddEnglishCounter.Call(q.h, uintptr(unsafe.Pointer(p)), 0, uintptr(unsafe.Pointer(&c))); r != 0 {
		return 0, errors.New("PdhAddEnglishCounter failed: " + path)
	}
	return c, nil
}

func (q *query) collect() {
	procPdhCollectQueryData.Call(q.h)
}

func (q *query) value(c uintptr) (float64, bool) {
	var v fmtValue
	r, _, _ := procPdhGetFormattedValue.Call(c, pdhFmtDouble|pdhFmtNoCap100, 0, uintptr(unsafe.Pointer(&v)))
	if r != 0 || v.CStatus > 1 {
		return 0, false
	}
	return v.Double, true
}

func (q *query) array(c uintptr) map[string]float64 {
	var size, count uint32
	r, _, _ := procPdhGetFormattedArray.Call(c, pdhFmtDouble|pdhFmtNoCap100, uintptr(unsafe.Pointer(&size)), uintptr(unsafe.Pointer(&count)), 0)
	if uint32(r) != pdhMoreData || size == 0 {
		return nil
	}
	buf := make([]byte, size)
	r, _, _ = procPdhGetFormattedArray.Call(c, pdhFmtDouble|pdhFmtNoCap100, uintptr(unsafe.Pointer(&size)), uintptr(unsafe.Pointer(&count)), uintptr(unsafe.Pointer(&buf[0])))
	if r != 0 {
		return nil
	}
	out := make(map[string]float64, count)
	items := unsafe.Slice((*fmtItem)(unsafe.Pointer(&buf[0])), count)
	for _, it := range items {
		if it.Value.CStatus <= 1 {
			out[windows.UTF16PtrToString(it.Name)] += it.Value.Double
		}
	}
	return out
}

func (q *query) close() {
	procPdhCloseQuery.Call(q.h)
}
