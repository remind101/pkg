// +build darwin
// +build !cgo

package cpu

import "github.com/remind101/pkg/reporter/hb2/internal/gopsutil/internal/common"

func perCPUTimes() ([]CPUTimesStat, error) {
	return []CPUTimesStat{}, common.NotImplementedError
}

func allCPUTimes() ([]CPUTimesStat, error) {
	return []CPUTimesStat{}, common.NotImplementedError
}
