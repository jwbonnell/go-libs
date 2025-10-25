
.PHONY: test test_pkg_db test_pkg_log bench_pkg_log

test: test_pkg_db test_pkg_log

test_pkg_db:
	go test -v ./pkg/db/...

test_pkg_log:
	go test -v ./pkg/log/...

bench_pkg_log:
	go test ./pkg/log/... -bench=. -benchmem

usage_basic:
	go run ./pkg/log/examples/basic_usage.go | jq