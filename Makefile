include Makeroutines.mk

# fix sirupsen/Sirupsen problem
define fix_sirupsen_case_sensitivity_problem
    @echo "# fixing sirupsen case sensitivity problem, please wait ..."
    @-rm -rf vendor/github.com/Sirupsen
    @-find ./ -type f -name "*.go" -exec sed -i -e 's/github.com\/Sirupsen\/logrus/github.com\/sirupsen\/logrus/g' {} \;
endef

# get required tools
get-tools:
	    @go get -u -f "github.com/alecthomas/gometalinter"
	    @gometalinter --install

# install dependencies
install-dep:
	$(call install_dependencies)
	$(fix_sirupsen_case_sensitivity_problem)

# update dependencies
update-dep:
	$(call update_dependencies)
	$(fix_sirupsen_case_sensitivity_problem)

# run checkstyle
checkstyle:
	    @echo "# running code analysis"
	    @gometalinter --vendor --exclude=vendor --deadline 1m --enable-gc --disable=aligncheck --disable=gotype --exclude=mock ./...
	    @echo "# done"

# run all tests
test:
	@echo "# running unit tests"
	@go test $$(go list ./... | grep -v /vendor/)

# get coverage percentage
coverage:
	@echo "# getting test coverage"
	@go test -cover $$(go list ./... | grep -v /vendor/)

# build all binaries
build:
	    @echo "# building"
	    @go build -a ./bgp/...
	    @echo "# done"

.PHONY: get-tools install-dep update-dep checkstyle build