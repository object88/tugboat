# Tugboat

Tugboat is a tool using to augment a [Helm](helm.sh) deployment by capturing the state of pods as their lifecycles end and start.

Requirements:
- pods must be labels with a release corresponding to the helm release name



The folder structure describes an intended for the future:
* `/apps`: home for individual PROGRAMs
* `/apps/PROGRAM`: top-level for a service, binary, or other non-package deliverable
* `/apps/PROGRAM/charts/PROGRAM`: (optional) Helm chart files
* `/apps/PROGRAM/main/main.go`: entry point
* `/apps/PROGRAM/cmd`: (optional) typical command structure
* `/apps/PROGRAM/pkg`: (optional) PROGRAM-specific packages
* `/apps/Dockerfile`: (optional) 
* `/bin`: (git-ignored) where PROGRAMs are compiled to
* `/build_tools`: 
* `/internal`: shared (non-PROGRAM-specific) packages for _internal use only_; cannot be used by outside repositories
* `/local`: (git-ignored) holds developer-owned scripts, tools, etc., for personal development
* `/mocks`: (git-ignored) holds mocks build by `go-mock` tooling to testing purposes
* `/pkg`: shared (non-PROGRAM-specific) packages; may be used by outside repositories
* `/testdata`: shared (non-PROGRAM-specific) test resources
* `/vendor`: vendored 3rd party packages
* `/build.sh`: script to build PROGRAMs
* `/docker.sh`: script to build docker images; invokes `/build_tools/Dockerfile`
* `/go.mod`, `/go.sum`: go vendoring files

## build.sh
The `build.sh` script looks for programs in the `apps` directory which have a `main/main.go` file, and builds them into the `bin` directory, using the current architecture and OS.  By default, it validates code quality by running Go tools such as test and vet.

### Usage:
```
./build.sh [FLAGS] [TARGET-1 [TARGET-2 [...]]]
```

### Options and flags
* `--fast`: alias for `--no-gen --no-test --no-verify --no-vet`
* `--no-gen`: do not call `go generate`; useful if mocked interfaces are unchanged
* `--no-test`: do not call `go test`
* `--no-verify`: do not call `go mod verify`
* `--no-vet`: do not call `go vet`

### Targets
If `TARGET-N` is specified, `build.sh` will look for a match in `/apps` and build it.

Targets are built in the order specified.

If no targets are specified, all PROGRAMs in `/apps` are built.  Targets are built in alphabetical order.

## docker.sh
The `docker.sh` script looks for programs in the `apps` directory and generates docker images for them.  It does this by passing options and targets to `build.sh` executed in `build_tools/Dockerfile`, which is named and tagged `gobuild:local`.  It is _strongly_ preferred that Dockerfiles for individual programs simply copy built binaries from that image into their own image.  This will greatly reduce the time to build programs.

Docker images are tagged using the current git tag and short-form SHA.

### Usage:
```
./docker.sh [FLAGS]  [TARGET-1 [TARGET-2 [...]]]
```

### Options and flags
* `--push`: push built docker images
* `--no-test`: pass flag into build.sh
* `--no-verify`: pass flag into build.sh
* `--no-vet`: pass flag into build.sh

### Targets
If `TARGET-N` is specified, `docker.sh` will look for a match in `/apps` and build it.  `docker.sh` will look for `Dockerfile` in the root of the program's directory.

Targets are built in the order specified.

If no targets are specified, all PROGRAMs in `/apps` are built.  Targets are built in alphabetical order.







Tugboat to deploy via helm
- custom resource to track release history
- when deploying, examine chart


Considerations:
- Tie into tekton pipeline