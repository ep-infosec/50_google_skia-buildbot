include ../make/bazel.mk

.PHONY: build
build: build-placeholders
	$(BAZEL) build --config=mayberemote //shaders/...

.PHONY: release
release: wasm_libs
	$(BAZEL) run --config=mayberemote //shaders:push_shaders_container
	$(MAKE) build-placeholders

.PHONY: wasm_libs
wasm_libs:
	docker pull gcr.io/skia-public/skia-wasm-release:prod
	rm -rf build/canvaskit
	mkdir -p build/canvaskit
	# Extract them from the most recently built docker container. This is built
	# with @skia//docker/skia-wasm-release/Dockerfile
	docker run -u $$(id -u ${USER}):$$(id -g ${USER}) --rm \
	    --volume `pwd`/build:/OUT gcr.io/skia-public/skia-wasm-release:prod \
	    sh -c "cp -r /tmp/canvaskit /tmp/VERSION /OUT/"
	echo "const SKIA_VERSION = '`cat build/VERSION`';" > build/version.js

.PHONY: push
push: release
	pushk shaders

bazel_release_ci:
	# COPY_FROM_DIR is expected to be set by the auto deploy mechanism, e.g.
	# https://github.com/google/skia/blob/44c81d1492738cb1c6ab3d68a925bd7e64201bca/infra/bots/task_drivers/push_bazel_apps_from_wasm_image/push_bazel_apps_from_wasm_image.go#L137
	cp -r $(COPY_FROM_DIR)/* build/
	echo "const SKIA_VERSION = '`cat ./build/VERSION`';" > ./build/version.js
	bazelisk run //shaders:push_shaders_container --workspace_status_command=bazel/override_container_label.sh
	$(MAKE) build-placeholders

# You can either run a local copy of scrapexchange, or port-forward the production instance:
#
#   kubectl port-forward service/scrapexchange 9000
.PHONY: run-local-instance
run-local-instance: build
	../_bazel_bin/shaders/go/shaders/shaders_/shaders --local \
		--resources_dir ../_bazel_bin/shaders/pages/development \
		--scrapexchange=http://localhost:9000 \
		--port=:8001 \
		--prom_port=:20001

.PHONY: run-with-custom
run-with-custom:
	echo "Using the wasm files in //shaders/wasm_libs/local_build and an in-memory scrapexchange"
	$(BAZEL) build --config=mayberemote //shaders/... --shaders_local_canvaskit
	../_bazel_bin/shaders/go/shaders/shaders_/shaders --local \
		--resources_dir ../_bazel_bin/shaders/pages/development \
		--fake_scrapexchange \
		--port=:8001 \
		--prom_port=:20001

.PHONY: build-placeholders
build-placeholders:
	rm -rf build
	mkdir -p build/canvaskit
	touch build/canvaskit/canvaskit.js
	touch build/canvaskit/canvaskit.wasm
	touch build/version.js
