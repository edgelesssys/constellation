cmake_minimum_required(VERSION 3.11)
project(constellation LANGUAGES C)

enable_testing()

# TODO(malt3): Remove this once every integration test is migrated to Bazel
add_test(NAME integration-csi COMMAND bash -c "go test -tags integration -c ./test/ && sudo ./test.test -test.v" WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/csi)
add_test(NAME integration-dm COMMAND bash -c "go test -tags integration -c ./test/ && sudo ./test.test -test.v" WORKING_DIRECTORY ${CMAKE_SOURCE_DIR}/disk-mapper/internal)
