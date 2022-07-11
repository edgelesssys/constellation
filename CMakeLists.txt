cmake_minimum_required(VERSION 3.11)
project(coordinator LANGUAGES C VERSION 0.1.0)

set(CLI_VERSION "v0.1.0" CACHE STRING "Version of CLI binary.")

enable_testing()

#
# core-os disk-mapper
#
add_custom_target(disk-mapper ALL
  DOCKER_BUILDKIT=1 docker build -o ${CMAKE_BINARY_DIR} -f Dockerfile.build --target disk-mapper .
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  BYPRODUCTS disk-mapper
)

#
# coordinator
#

add_custom_target(coordinator ALL
  DOCKER_BUILDKIT=1 docker build -o ${CMAKE_BINARY_DIR} --build-arg PROJECT_VERSION=${PROJECT_VERSION} -f Dockerfile.build --target coordinator .
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  BYPRODUCTS coordinator
)


#
# cli
#

add_custom_target(cli ALL
  CGO_ENABLED=0 go build -o ${CMAKE_BINARY_DIR}/constellation -tags=gcp -ldflags "-buildid='' -X github.com/edgelesssys/constellation/internal/constants.VersionInfo=${CLI_VERSION}"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/cli
  BYPRODUCTS constellation
)

#
# testing / debugging
#

add_custom_target(debug_coordinator
  go build -o ${CMAKE_BINARY_DIR}/debug_coordinator -buildvcs=false -ldflags "-buildid='' -X main.version=${PROJECT_VERSION}"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/coordinator/cmd/coordinator
  BYPRODUCTS debug_coordinator
)

#
# debugd
#

add_custom_target(debugd ALL
  CGO_ENABLED=0 go build -o ${CMAKE_BINARY_DIR}/debugd -buildvcs=false -ldflags "-buildid=''"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/debugd/debugd/cmd/debugd
  BYPRODUCTS debugd
)

#
# cdbg
#

add_custom_target(cdbg ALL
  CGO_ENABLED=0 go build -o ${CMAKE_BINARY_DIR}/cdbg -buildvcs=false -ldflags "-buildid=''"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/debugd/cdbg
  BYPRODUCTS cdbg
)

add_test(NAME unit-main COMMAND go test -race -count=3 ./... WORKING_DIRECTORY ${CMAKE_SOURCE_DIR})
add_test(NAME unit-hack COMMAND go test -race -count=3 ./... WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/hack)
add_test(NAME integration-coord COMMAND go test -v -tags integration ./test/  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR})
add_test(NAME integration-mount COMMAND bash -c "go test -tags integration -c ./test/ && sudo ./test.test -test.v -v 9"  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/mount)
add_test(NAME integration-dm COMMAND bash -c "go test -tags integration -c ./test/ && sudo ./test.test -test.v"  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/state)
add_test(NAME integration-etcd COMMAND go test -v --race -cover -count=3 -tags integration WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/coordinator/store/)
