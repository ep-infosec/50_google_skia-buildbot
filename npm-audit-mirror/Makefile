include ../make/bazel.mk

generate:
	go generate ./go/...

testgo:
	go test ./go/...

run-local-instance: build
	../_bazel_bin/npm-audit-mirror/go/npm-audit-mirror/npm-audit-mirror_/npm-audit-mirror --resources_dir ../_bazel_bin/npm-audit-mirror/pages/development --fs_namespace=npm-audit-mirror-staging --workdir /tmp/ --local

build: generate
	$(BAZEL) build --config=mayberemote //npm-audit-mirror/...

release: build
	$(BAZEL) run --config=mayberemote //npm-audit-mirror:push_npm_audit_mirror_container

push: release
	$(BAZEL) run --config=mayberemote //kube/go/pushk -- --use-temp-checkout npm-audit-mirror
