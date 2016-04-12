// +build windows

package load

import (
	"github.com/remind101/pkg/reporter/hb2/internal/gopsutil/internal/common"
)

func LoadAvg() (*LoadAvgStat, error) {
	ret := LoadAvgStat{}

	return &ret, common.NotImplementedError
}

func Misc() (*MiscStat, error) {
	ret := MiscStat{}

	return &ret, common.NotImplementedError
}
