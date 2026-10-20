package util

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/mndrix/tap-go"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
	"github.com/opencontainers/runtime-tools/cgroups"
)

const (
	// CPUCgroupPrefix is default path prefix where CPU cgroups are created
	CPUCgroupPrefix string = "/sys/fs/cgroup/cpu,cpuacct"
)

// ValidateLinuxResourcesCPU validates if Linux.Resources.CPU is set to
// correct values, the same as given values in the config.
func ValidateLinuxResourcesCPU(config *rspec.Spec, t *tap.T, state *rspec.State) error {
	cg, err := cgroups.FindCgroup()
	t.Ok((err == nil), "find cpu cgroup")
	if err != nil {
		t.Diagnostic(err.Error())
		return nil
	}

	lcd, err := cg.GetCPUData(state.Pid, config.Linux.CgroupsPath)
	t.Ok((err == nil), "get cpu cgroup data")
	if err != nil {
		t.Diagnostic(err.Error())
		return nil
	}

	checkValue(t, "cpu shares", config.Linux.Resources.CPU.Shares, lcd.Shares)
	checkValue(t, "cpu period", config.Linux.Resources.CPU.Period, lcd.Period)
	checkValue(t, "cpu quota", config.Linux.Resources.CPU.Quota, lcd.Quota)

	t.Ok(lcd.Cpus == config.Linux.Resources.CPU.Cpus, "cpu cpus is set correctly")
	t.Diagnosticf("expect: %s, actual: %s", config.Linux.Resources.CPU.Cpus, lcd.Cpus)

	t.Ok(lcd.Mems == config.Linux.Resources.CPU.Mems, "cpu mems is set correctly")
	t.Diagnosticf("expect: %s, actual: %s", config.Linux.Resources.CPU.Mems, lcd.Mems)

	return nil
}

// ValidateLinuxResourcesCPUEmpty validates Linux.Resources.CPU is set to
// correct values, when each value are set to the default ones.
func ValidateLinuxResourcesCPUEmpty(config *rspec.Spec, t *tap.T, state *rspec.State) error {
	outShares, err := os.ReadFile(filepath.Join(CPUCgroupPrefix, "cpu.shares"))
	if err != nil {
		return nil
	}
	sh, _ := strconv.Atoi(strings.TrimSpace(string(outShares)))
	defaultShares := uint64(sh)

	outPeriod, err := os.ReadFile(filepath.Join(CPUCgroupPrefix, "cpu.cfs_period_us"))
	if err != nil {
		return nil
	}
	pe, _ := strconv.Atoi(strings.TrimSpace(string(outPeriod)))
	defaultPeriod := uint64(pe)

	outQuota, err := os.ReadFile(filepath.Join(CPUCgroupPrefix, "cpu.cfs_quota_us"))
	if err != nil {
		return nil
	}
	qu, _ := strconv.Atoi(strings.TrimSpace(string(outQuota)))
	defaultQuota := int64(qu)

	defaultCpus := fmt.Sprintf("0-%d", runtime.NumCPU()-1)
	defaultMems := "0"

	cg, err := cgroups.FindCgroup()
	t.Ok((err == nil), "find cpu cgroup")
	if err != nil {
		t.Diagnostic(err.Error())
		return nil
	}

	lcd, err := cg.GetCPUData(state.Pid, config.Linux.CgroupsPath)
	t.Ok((err == nil), "get cpu cgroup data")
	if err != nil {
		t.Diagnostic(err.Error())
		return nil
	}

	checkValue(t, "cpu shares", &defaultShares, lcd.Shares)
	checkValue(t, "cpu period", &defaultPeriod, lcd.Period)
	checkValue(t, "cpu quota", &defaultQuota, lcd.Quota)

	t.Ok(lcd.Cpus == defaultCpus, "cpu cpus is set correctly")
	t.Diagnosticf("expect: %s, actual: %s", defaultCpus, lcd.Cpus)

	t.Ok(lcd.Mems == defaultMems, "cpu mems is set correctly")
	t.Diagnosticf("expect: %s, actual: %s", defaultMems, lcd.Mems)

	return nil
}
