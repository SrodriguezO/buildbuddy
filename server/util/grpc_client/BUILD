load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["grpc_client.go"],
    importpath = "github.com/buildbuddy-io/buildbuddy/server/util/grpc_client",
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_grpc_ecosystem_go_grpc_prometheus//:go_default_library",
        "@org_golang_google_grpc//:go_default_library",
        "@org_golang_google_grpc//credentials/google:go_default_library",
    ],
)
