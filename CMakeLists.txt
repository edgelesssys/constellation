cmake_minimum_required(VERSION 3.11)
project(coordinator LANGUAGES C VERSION 0.1.0)

enable_testing()

#
# core-os disk-mapper
#
add_custom_target(disk-mapper
  go build -o ${CMAKE_BINARY_DIR}/disk-mapper -ldflags "-s -w"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/state/cmd
)

#
# coordinator
#

add_custom_target(coordinator ALL
  go build -o ${CMAKE_BINARY_DIR} -tags=gcp -buildvcs=false -ldflags "-s -w -buildid='' -X main.version=${PROJECT_VERSION}"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/coordinator/cmd/coordinator)


#
# cli
#

add_custom_target(cli ALL
  go build -o ${CMAKE_BINARY_DIR}/constellation -buildvcs=false -tags=gcp -ldflags "-buildid='' -X github.com/edgelesssys/constellation/internal/constants.CliVersion=${PROJECT_VERSION}"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/cli)

#
# testing / debugging
#

add_custom_target(debug_coordinator ALL
  go build -o ${CMAKE_BINARY_DIR}/debug_coordinator -buildvcs=false -ldflags "-buildid='' -X main.version=${PROJECT_VERSION}"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/coordinator/cmd/coordinator)

add_test(NAME unittest COMMAND go test -race -count=3 ./... WORKING_DIRECTORY ${CMAKE_SOURCE_DIR})
add_test(NAME integrationtest COMMAND go test -v -tags integration ./test/  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR})
add_test(NAME etcd-unittest COMMAND go test -v --race -cover -count=3 -tags integration WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/coordinator/store/)
