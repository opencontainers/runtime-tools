package translate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/codegangsta/cli"
	rspec "github.com/opencontainers/runtime-spec/specs-go"
)

func FromContainer(data interface{}, context *cli.Context) (translated interface{}, err error) {
	dataMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("data is not a map[string]interface{}: %s", data)
	}

	linuxInterface, ok := dataMap["linux"]
	if !ok {
		return data, nil
	}

	linux, ok := linuxInterface.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("data.linux is not a map[string]interface{}: %s", linuxInterface)
	}

	namespacesInterface, ok := linux["namespaces"]
	if !ok {
		return data, nil
	}

	namespaces, ok := namespacesInterface.([]interface{})
	if !ok {
		return nil, fmt.Errorf("data.linux.namespaces is not an array: %s", namespacesInterface)
	}

	for index, namespaceInterface := range namespaces {
		namespace, ok := namespaceInterface.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("data.linux.namespaces[%d] is not a map[string]interface{}: %s", index, namespaceInterface)
		}
		err := namespaceFromContainer(&namespace, index, context)
		if err != nil {
			return nil, err
		}
	}

	return data, nil
}

func namespaceFromContainer(namespace *map[string]interface{}, index int, context *cli.Context) (err error) {
	fromContainerInterface, ok := (*namespace)["fromContainer"]
	if ok {
		fromContainer, ok := fromContainerInterface.(string)
		if !ok {
			return fmt.Errorf("data.linux.namespaces[%d].fromContainer is not a string: %s", index, fromContainerInterface)
		}
		delete(*namespace, "fromContainer")
		runtime := context.String("runtime")
		if (len(runtime) == 0) {
			return fmt.Errorf("translating fromContainer requires a non-empty --runtime")
		}
		command := exec.Command(runtime, "state", fromContainer)
		var out bytes.Buffer
		command.Stdout = &out
		err := command.Run()
		if err != nil {
			return err
		}
		var state rspec.State
		err = json.Unmarshal(out.Bytes(), &state)
		if err != nil {
			return err
		}
		namespaceTypeInterface, ok := (*namespace)["type"]
		if !ok {
			return fmt.Errorf("data.linux.namespaces[%d].type is missing: %s", index, fromContainerInterface)
		}
		namespaceType, ok := namespaceTypeInterface.(string)
		if !ok {
			return fmt.Errorf("data.linux.namespaces[%d].type is not a string: %s", index, namespaceTypeInterface)
		}
		switch namespaceType {
		case "network": namespaceType = "net"
		case "mount": namespaceType = "mnt"
		}
		proc := "/proc"  // FIXME: lookup in /proc/self/mounts, check right namespace
		path := filepath.Join(proc, fmt.Sprint(state.Pid), "ns", namespaceType)
		(*namespace)["path"] = path
	}
	return nil
}
