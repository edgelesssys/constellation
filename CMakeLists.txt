cmake_minimum_required(VERSION 3.11)
project(constellation LANGUAGES C)
set(CLI_BUILD_TAGS "" CACHE STRING "Tags passed to go build of Constellation CLI.")
set(BAZEL "bazel" CACHE STRING "Path to bazel binary.")

enable_testing()

#
# disk-mapper
#
add_custom_target(disk-mapper ALL
  COMMAND ${BAZEL} build //disk-mapper/cmd:cmd
  COMMAND cp \$$\(${BAZEL} cquery --output=files //disk-mapper/cmd:cmd\) ${CMAKE_BINARY_DIR}/disk-mapper
  COMMAND chmod +w ${CMAKE_BINARY_DIR}/disk-mapper
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  BYPRODUCTS disk-mapper
)

#
# measurement-reader
#
add_custom_target(measurement-reader ALL
  COMMAND ${BAZEL} build //measurement-reader/cmd:cmd
  COMMAND cp \$$\(${BAZEL} cquery --output=files //measurement-reader/cmd:cmd\) ${CMAKE_BINARY_DIR}/measurement-reader
  COMMAND chmod +w ${CMAKE_BINARY_DIR}/measurement-reader
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  BYPRODUCTS measurement-reader
)

#
# bootstrapper
#
add_custom_target(bootstrapper ALL
  COMMAND ${BAZEL} build //bootstrapper/cmd/bootstrapper:bootstrapper
  COMMAND cp \$$\(${BAZEL} cquery --output=files //bootstrapper/cmd/bootstrapper:bootstrapper\) ${CMAKE_BINARY_DIR}/bootstrapper
  COMMAND chmod +w ${CMAKE_BINARY_DIR}/bootstrapper
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  BYPRODUCTS bootstrapper
)

#
# upgrade-agent
#
add_custom_target(upgrade-agent ALL
  COMMAND ${BAZEL} build //upgrade-agent/cmd:cmd
  COMMAND cp \$$\(${BAZEL} cquery --output=files //upgrade-agent/cmd:cmd\) ${CMAKE_BINARY_DIR}/upgrade-agent
  COMMAND chmod +w ${CMAKE_BINARY_DIR}/upgrade-agent
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  BYPRODUCTS upgrade-agent
)

#
# cli
#
add_custom_target(cli ALL
  COMMAND ${BAZEL} build --@io_bazel_rules_go//go/config:tags='${CLI_BUILD_TAGS}' --platforms=@local_config_platform//:host //cli:cli_oss
  COMMAND cp \$$\(${BAZEL} cquery --@io_bazel_rules_go//go/config:tags='${CLI_BUILD_TAGS}' --platforms=@local_config_platform//:host --output=files //cli:cli_oss\) ${CMAKE_BINARY_DIR}/constellation
  COMMAND chmod +w ${CMAKE_BINARY_DIR}/constellation
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  BYPRODUCTS constellation
)

#
# debugd
#
add_custom_target(debugd ALL
  COMMAND ${BAZEL} build //debugd/cmd/debugd:debugd
  COMMAND cp \$$\(${BAZEL} cquery --output=files //debugd/cmd/debugd:debugd\) ${CMAKE_BINARY_DIR}/debugd
  COMMAND chmod +w ${CMAKE_BINARY_DIR}/debugd
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  BYPRODUCTS debugd
)

#
# cdbg
#
add_custom_target(cdbg ALL
  COMMAND ${BAZEL} build --platforms=@local_config_platform//:host //debugd/cmd/cdbg:cdbg
  COMMAND cp \$$\(${BAZEL} cquery --platforms=@local_config_platform//:host --output=files //debugd/cmd/cdbg:cdbg\) ${CMAKE_BINARY_DIR}/cdbg
  COMMAND chmod +w ${CMAKE_BINARY_DIR}/cdbg
  WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}
  BYPRODUCTS cdbg
)

add_test(NAME unit COMMAND bash -c "go test -race -count=3 $(go list -f '{{.Dir}}/...' -m | xargs)" WORKING_DIRECTORY ${CMAKE_SOURCE_DIR})
add_test(NAME integration-node-operator COMMAND make test WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/operators/constellation-node-operator)
add_test(NAME integration-csi COMMAND bash -c "go test -tags integration -c ./test/ && sudo ./test.test -test.v" WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/csi)
add_test(NAME integration-dm COMMAND bash -c "go test -tags integration -c ./test/ && sudo ./test.test -test.v" WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/disk-mapper/internal)
add_test(NAME integration-license COMMAND bash -c "go test -tags integration" WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/internal/license)
