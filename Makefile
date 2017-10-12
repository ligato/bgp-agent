include Makeroutines.mk

COVER_DIR=/tmp/

# fix sirupsen/Sirupsen problem
define fix_sirupsen_case_sensitivity_problem
    @echo "# fixing sirupsen case sensitivity problem, please wait ..."
    @-rm -rf vendor/github.com/Sirupsen
    @-find ./ -type f -name "*.go" -exec sed -i -e 's/github.com\/Sirupsen\/logrus/github.com\/sirupsen\/logrus/g' {} \;
endef


# run all tests with coverage
define test_cover_only
	@echo "# running unit tests with coverage analysis"
	@go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit1.out ./bgp
	@go test -covermode=count -coverprofile=${COVER_DIR}coverage_unit2.out ./bgp/gobgp
	@echo "# merging coverage results"
    @gocovmerge ${COVER_DIR}coverage_unit1.out ${COVER_DIR}coverage_unit2.out  > ${COVER_DIR}coverage.out
    @echo "# coverage data generated into ${COVER_DIR}coverage.out"
    @echo "# done"
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
	    @go get -u -f "github.com/wadey/gocovmerge"

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
	    @gometalinter --vendor --exclude=vendor --deadline 1m --enable-gc --disable=aligncheck --disable=gotype --disable=gotypex --exclude=mock ./...
	    @echo "# done"

# run all tests
test:
	@echo "# running unit tests"
	@go test $$(go list ./... | grep -v /vendor/)

# run tests with coverage report
test-cover:
	$(call test_cover_only)

# get coverage percentage in console(without report)
test-cover-without-report:
	@echo "# getting test coverage"
	@go test -cover $$(go list ./... | grep -v /vendor/)

# build examples
build-examples:
		@echo "# building plugin examples"
		@cd examples/gobgp_watch_plugin && go build

# run examples
run-examples:
	    @make build-examples
	    @echo "# running examples"
	    @./scripts/run_gobgp_watcher_examples.sh
	    @echo "# done"

# run checkstyle
clean:
	    @echo "# cleaning"
		@rm -f examples/gobgp_watch_plugin/gobgp_watch_plugin
		@rm -f docker/gobgp_route_reflector/gobgp-client-in-host/gobgp-client-in-host
		@rm -f docker/gobgp_route_reflector/usage_scripts/gobgp-client-in-host/log
		@rm -f log
		@echo "# done"

#run all
all:
	    @echo "# running all"
	    @make get-tools
	    @make install-dep
	    @make update-dep
	    @make checkstyle
	    @make build
	    @make test-cover
	    @make build-examples
	    @make run-examples
	    @make clean

.PHONY: build get-tools install-dep update-dep checkstyle test test-cover test-cover-without-report build-examples run-examples clean all