include Makeroutines.mk

# get required tools
get-tools:
	    @go get -u -f "github.com/alecthomas/gometalinter"
	    @gometalinter --install

# install dependencies
install-dep:
	$(call install_dependencies)

# update dependencies
update-dep:
	$(call update_dependencies)

# run checkstyle
checkstyle:
	    @echo "# running code analysis"
	    @gometalinter --vendor --exclude=vendor --deadline 1m --enable-gc --disable=aligncheck --disable=gotype --exclude=mock ./...
	    @echo "# done"

# build all binaries
build:
	    @echo "# building"
	    @go build -a ./bgp/...
	    @echo "# done"

.PHONY: get-tools install-dep update-dep checkstyle build