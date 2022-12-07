cmake_minimum_required(VERSION 3.11)
project(constellation LANGUAGES C VERSION 2.2.2)
set(CLI_BUILD_TAGS "" CACHE STRING "Tags passed to go build of Constellation CLI.")

enable_testing()

#
# core-os disk-mapper
#
add_custom_target(disk-mapper ALL
  DOCKER_BUILDKIT=1 docker build -o ${CMAKE_BINARY_DIR} --build-arg PROJECT_VERSION=${PROJECT_VERSION} -f Dockerfile.build --target disk-mapper .
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  BYPRODUCTS disk-mapper
)

#
# bootstrapper
#
add_custom_target(bootstrapper ALL
  DOCKER_BUILDKIT=1 docker build -o ${CMAKE_BINARY_DIR} --build-arg PROJECT_VERSION=${PROJECT_VERSION} -f Dockerfile.build --target bootstrapper .
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  BYPRODUCTS bootstrapper
)

#
# upgrade-agent
#
add_custom_target(upgrade-agent ALL
  DOCKER_BUILDKIT=1 docker build -o ${CMAKE_BINARY_DIR} --build-arg PROJECT_VERSION=${PROJECT_VERSION} -f Dockerfile.build --target upgrade-agent .
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  BYPRODUCTS upgrade-agent
)

#
# cli
#
add_custom_target(cli ALL
  CGO_ENABLED=0 go build -o ${CMAKE_BINARY_DIR}/constellation -tags='${CLI_BUILD_TAGS}' -ldflags "-buildid='' -X github.com/edgelesssys/constellation/v2/internal/constants.VersionInfo=${PROJECT_VERSION}"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/cli
  BYPRODUCTS constellation
)

#
# testing / debugging
#
add_custom_target(debug_bootstrapper
  go build -o ${CMAKE_BINARY_DIR}/debug_bootstrapper -buildvcs=false -ldflags "-buildid='' -X main.version=${PROJECT_VERSION}"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/bootstrapper/cmd/bootstrapper
  BYPRODUCTS debug_bootstrapper
)

#
# debugd
#
add_custom_target(debugd ALL
  CGO_ENABLED=0 go build -o ${CMAKE_BINARY_DIR}/debugd -buildvcs=false -ldflags "-buildid=''"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/debugd/cmd/debugd
  BYPRODUCTS debugd
)

#
# cdbg
#
add_custom_target(cdbg ALL
  CGO_ENABLED=0 go build -o ${CMAKE_BINARY_DIR}/cdbg -buildvcs=false -ldflags "-buildid=''"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/debugd/cmd/cdbg
  BYPRODUCTS cdbg
)

add_test(NAME unit COMMAND bash -c "go test -race -count=3 $(go list -f '{{.Dir}}/...' -m | xargs)" WORKING_DIRECTORY ${CMAKE_SOURCE_DIR})
add_test(NAME integration-node-operator COMMAND make test WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/operators/constellation-node-operator)
add_test(NAME integration-csi COMMAND bash -c "go test -tags integration -c ./test/ && sudo ./test.test -test.v" WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/csi)
add_test(NAME integration-dm COMMAND bash -c "go test -tags integration -c ./test/ && sudo ./test.test -test.v" WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/disk-mapper/internal)
add_test(NAME integration-license COMMAND bash -c "go test -tags integration" WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/internal/license)

#
# format
#
add_custom_target(format
  COMMAND bash -c "shfmt -i 2 -s -w -sr $(shfmt -f ${CMAKE_SOURCE_DIR} | grep -v helm/charts/cilium)"
  COMMAND bash -c "gofumpt -l -w ${CMAKE_SOURCE_DIR}"
  VERBATIM
)
