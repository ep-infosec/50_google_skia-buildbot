include ../make/bazel.mk

generate:
	go generate ./go/...

testgo:
	go test ./go/...

serve:
	@echo "Bazel doesn't have a serve-all option."
	@echo "For the module under development, try running something like:"
	@echo "bazel run modules/bugs-central-scaffold-sk:demo_page_server"

run-local-instance: build
	# Run "bazel build //bugs-central/..." after making any frontend changes
	# to pick up the change without restarting the golang server.
	../_bazel_bin/bugs-central/go/bugs-central/bugs-central_/bugs-central --local --resources_dir ../_bazel_bin/bugs-central/pages/development --fs_namespace=bugs-central-staging

testjs:
	$(BAZEL) test --config=mayberemote modules/...

build: generate
	$(BAZEL) build --config=mayberemote go/bugs-central/... modules/... pages/...

release: build
	$(BAZEL) run --config=mayberemote //bugs-central:push_bugs_central_container

push: release
	$(BAZEL) run --config=mayberemote //kube/go/pushk -- --use-temp-checkout bugs-central

