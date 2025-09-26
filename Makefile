GO ?= go
TESTFOLDER := $(shell $(GO) list ./... | grep -E 'gin$$|binding$$|render$$' | grep -v examples)
TESTTAGS ?= ""

.PHONY: test
# Run tests to verify code functionality.
test:
	echo "mode: count" > coverage.out
	for d in $(TESTFOLDER); do \
		$(GO) test $(TESTTAGS) -v -covermode=count -coverprofile=profile.out $$d > tmp.out; \
		cat tmp.out; \
		if grep -q "^--- FAIL" tmp.out; then \
			rm tmp.out; \
			exit 1; \
		elif grep -q "build failed" tmp.out; then \
			rm tmp.out; \
			exit 1; \
		elif grep -q "setup failed" tmp.out; then \
			rm tmp.out; \
			exit 1; \
		fi; \
		if [ -f profile.out ]; then \
			cat profile.out | grep -v "mode:" >> coverage.out; \
			rm profile.out; \
		fi; \
	done