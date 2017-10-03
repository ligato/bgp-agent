include Makeroutines.mk

# fix sirupsen/Sirupsen problem
define fix_sirupsen_case_sensitivity_problem
    @echo "# fixing sirupsen case sensitivity problem, please wait ..."
    @-rm -rf vendor/github.com/Sirupsen
    @-find ./ -type f -name "*.go" -exec sed -i -e 's/github.com\/Sirupsen\/logrus/github.com\/sirupsen\/logrus/g' {} \;
endef

# build all binaries
build:
	    @echo "# building"
	    @go build -a ./bgp/...
	    @echo "# done"

# get required tools
get-tools:
	    @go get -u -f "github.com/alecthomas/gometalinter"
	    @gometalinter --install

# install dependencies
install-dep:
	@echo "# install dependencies"
	$(call install_dependencies)
	$(fix_sirupsen_case_sensitivity_problem)

# update dependencies
update-dep:
	@echo "# update dependencies"
	$(call update_dependencies)
	$(fix_sirupsen_case_sensitivity_problem)

# run checkstyle
checkstyle:
	    @echo "# running code analysis"
	    @gometalinter --vendor --exclude=vendor --deadline 1m --enable-gc --disable=aligncheck --disable=gotype --exclude=mock ./...
	    @echo "# done"

# build examples
build-examples:
		@echo "# building plugin examples"
		@cd examples/gobgp_watch_plugin && go build

# run checkstyle
run-examples:
	    @echo "# running examples"
	    @make build-examples
	    @./scripts/run_gobgp_watcher_examples.sh
	    @echo "# done"

# run checkstyle
clean:
	    @echo "# cleaning"
		@rm -f examples/gobgp_watch_plugin/gobgp_watch_plugin
		@rm -f docker/gobgp_route_reflector/gobgp-client-in-host/gobgp-client-in-host
		@rm -f docker/gobgp_route_reflector/gobgp-client-in-host/log
		@rm -f docker/gobgp_route_reflector/gobgp-client-in-docker/gobgp-client-in-docker
		@rm -f docker/gobgp_route_reflector/gobgp-client-in-docker/log
		@rm -f docker/gobgp_route_reflector/gobgp-benchmark/gobgp-benchmark
		@rm -f docker/gobgp_route_reflector/gobgp-benchmark/log
		@echo "# done"

#run all
all:
	    @echo "# running all"
	    @make get-tools
	    @make install-dep
	    @make update-dep
	    @make checkstyle
	    @make build
	    @make build-examples
	    @make run-examples
	    @make clean

.PHONY: build install-dep update-dep checkstyle coverage clean all run-examples