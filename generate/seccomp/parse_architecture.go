package seccomp

import (
	"fmt"

	types "github.com/opencontainers/runtime-spec/specs-go"
)

// ParseArchitectureFlag takes the raw string passed with the --arch flag, parses it
// and updates the Seccomp config accordingly
func ParseArchitectureFlag(architectureArgs []string, config *types.Seccomp) error {

	var arches []types.Arch
	for _, arg := range architectureArgs {
		correctedArch, err := parseArch(arg)
		if err != nil {
			return err
		}
		shouldAppend := true
		for _, alreadySpecified := range config.Architectures {
			if correctedArch == alreadySpecified {
				shouldAppend = false
			}
		}
		if shouldAppend {
			arches = append(arches, correctedArch)
			config.Architectures = arches
		}
	}
	return nil
}

func parseArch(arch string) (types.Arch, error) {
	arches := map[string]types.Arch{
		"x86":         types.ArchX86,
		"amd64":       types.ArchX86_64,
		"x32":         types.ArchX32,
		"arm":         types.ArchARM,
		"arm64":       types.ArchAARCH64,
		"mips":        types.ArchMIPS,
		"mips64":      types.ArchMIPS64,
		"mips64n32":   types.ArchMIPS64N32,
		"mipsel":      types.ArchMIPSEL,
		"mipsel64":    types.ArchMIPSEL64,
		"mipsel64n32": types.ArchMIPSEL64N32,
		"ppc":         types.ArchPPC,
		"ppc64":       types.ArchPPC64,
		"ppc64le":     types.ArchPPC64LE,
		"s390":        types.ArchS390,
		"s390x":       types.ArchS390X,
	}
	a, ok := arches[arch]
	if !ok {
		return "", fmt.Errorf("Unrecognized architecutre: %s", arch)
	}
	return a, nil
}
