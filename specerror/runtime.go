package specerror

import (
	"fmt"

	rfc2119 "github.com/opencontainers/runtime-tools/error"
)

// define error codes
const (
	// CreateWithBundlePathAndID represents "This operation MUST [generate an error](#errors) if it is not provided a path to the bundle and the container ID to associate with the container."
	CreateWithBundlePathAndID = "This operation MUST [generate an error](#errors) if it is not provided a path to the bundle and the container ID to associate with the container."
	// CreateWithUniqueID represents "If the ID provided is not unique across all containers within the scope of the runtime, or is not valid in any other way, the implementation MUST [generate an error](#errors) and a new container MUST NOT be created."
	CreateWithUniqueID = "If the ID provided is not unique across all containers within the scope of the runtime, or is not valid in any other way, the implementation MUST [generate an error](#errors) and a new container MUST NOT be created."
	// CreateNewContainer represents "This operation MUST create a new container."
	CreateNewContainer = "This operation MUST create a new container."
)

var (
	createRef = func(version string) (reference string, err error) {
		return fmt.Sprintf(referenceTemplate, version, "runtime.md#create"), nil
	}
)

func init() {
	register(CreateWithBundlePathAndID, rfc2119.Must, createRef)
	register(CreateWithUniqueID, rfc2119.Must, createRef)
	register(CreateNewContainer, rfc2119.Must, createRef)
}
