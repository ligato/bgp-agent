# get required tools
get-tools:
	    @go get -u -f "github.com/alecthomas/gometalinter"
	    @go install "github.com/alecthomas/gometalinter"

# run checkstyle
checkstyle:
	    @echo "# running code analysis"
	    @gometalinter --vendor --exclude=vendor --deadline 1m --enable-gc --disable=aligncheck --disable=gotype --exclude=mock ./...
	    @echo "# done"

.PHONY: get-tools checkstyle
