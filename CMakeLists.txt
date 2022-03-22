cmake_minimum_required(VERSION 3.11)
project(coordinator LANGUAGES C VERSION 0.1.0)

enable_testing()
option(COORDINATOR_STATIC_MUSL "use musl and compile coordinator statically")

if(NOT CMAKE_BUILD_TYPE)
  set(CMAKE_BUILD_TYPE Debug)
endif()
if(CMAKE_BUILD_TYPE STREQUAL Debug)
  set(CARGOTARGET debug)
else()
  set(CARGOTARGET release)
  set(CARGOFLAGS --release)
endif()

if(COORDINATOR_STATIC_MUSL)
  set(RUST_STATICLIB_LDFLAGS -static ${RUST_STATICLIB_LDFLAGS})
  set(RUSTTARGETTRIPLE x86_64-unknown-linux-musl)
  set(CARGOFLAGS ${CARGOFLAGS} "--target=${RUSTTARGETTRIPLE}")
  set(CARGOTARGET ${RUSTTARGETTRIPLE}/${CARGOTARGET})
else()
   set(RUST_STATICLIB_LDFLAGS -ldl -lm -lrt ${RUST_STATICLIB_LDFLAGS})
endif()

set(NITRO_CFLAGS '-I${CMAKE_BINARY_DIR}/nitro/${CARGOTARGET} -I${CMAKE_BINARY_DIR}/nitro/${CARGOTARGET}/headers')
set(NITRO_LDFLAGS '${CMAKE_BINARY_DIR}/nitro/${CARGOTARGET}/libnitro.a ${RUST_STATICLIB_LDFLAGS}')

#
# coordinator
#

add_custom_target(nitro
  CARGO_TARGET_DIR=${CMAKE_BINARY_DIR}/nitro cargo build ${CARGOFLAGS}
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/3rdparty/aws-nitro-enclaves-ffi)

add_custom_target(coordinator ALL
  ${CMAKE_COMMAND} -E env CGO_CFLAGS=${NITRO_CFLAGS}
  ${CMAKE_COMMAND} -E env CGO_LDFLAGS=${NITRO_LDFLAGS}
  go build -o ${CMAKE_BINARY_DIR} -tags=aws,gcp -buildvcs=false -ldflags "-buildid='' -X main.version=${PROJECT_VERSION}"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/coordinator/cmd/coordinator)

add_dependencies(coordinator nitro)

#
# cli
#

add_custom_target(cli ALL
  ${CMAKE_COMMAND} -E env CGO_CFLAGS=${NITRO_CFLAGS}
  ${CMAKE_COMMAND} -E env CGO_LDFLAGS=${NITRO_LDFLAGS}
  go build -o ${CMAKE_BINARY_DIR}/constellation -buildvcs=false -tags=aws,gcp -ldflags "-buildid='' -X github.com/edgelesssys/constellation/cli/defaults.Version=${PROJECT_VERSION}"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/cli)

add_dependencies(cli nitro)

#
# testing / debugging
#

add_custom_target(debug_coordinator ALL
  go build -o ${CMAKE_BINARY_DIR}/debug_coordinator -buildvcs=false -ldflags "-buildid='' -X main.version=${PROJECT_VERSION}"
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/coordinator/cmd/coordinator)

add_test(NAME unittest COMMAND go test -race -count=3 ./... WORKING_DIRECTORY ${CMAKE_SOURCE_DIR})
add_test(NAME integrationtest COMMAND go test -v -tags integration ./test/  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR})
add_test(NAME etcd-unittest COMMAND go test -v --race -cover -count=3 -tags integration WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/coordinator/store/)
