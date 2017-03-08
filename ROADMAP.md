# OCI runtime-tools Roadmap

This document serves to provide a long term roadmap on our quest to a 1.0 version of the OCI runtime-tools.
Its goal is to help both maintainers and contributors to find meaningful tasks to focus on and create a low noise environment.
The items in the 1.0 roadmap can be broken down into smaller milestones that are easy to accomplish.
The topics below are broad and small working groups will be needed for each to define scope and requirements or if the feature is required at all for the OCI level.
Topics listed in the roadmap do not mean that they will be implemented or added but are areas that need discussion to see if they fit in to the goals of the OCI.

Listed topics may defer to the [project wiki][runtime-wiki] for collaboration.

## 1.0

### OCI runtime spec configuration files Generation

Generate configuration JSON for an OCI bundle.

All fields defined in [Container Configuration file][runtime-sepc-config] should be support by oci-runtime-generate.

*Owner:*

### Bundle Validation

Validate an OCI runtime bundle.

Bundle structure and all fileds in config.json should and can be validated by oci-runtime-validate.

*Owner:*

### Runtime Validation

Validate a runtime whether meets requirments of OCI runtime specs.

runtime-tools should and can validate a runtime from aspects of contianer lifecycle, operations  and application of [container configuration files][runtime-spec-config] (like cgroups setting, devices setting, etc). 

*Owner:*


[runtime-wiki]: https://github.com/opencontainers/runtime-tools/wiki/RoadMap
[runtime-spec-config]: https://github.com/opencontainers/runtime-spec/blob/master/config.md 
