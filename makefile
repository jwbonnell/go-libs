
COLORIZE := sed -E \
  -e 's/(PASS)/\x1b[32m\1\x1b[0m/g' \
  -e 's/(FAIL)/\x1b[31m\1\x1b[0m/g' \
  -e 's/(SKIP)/\x1b[33m\1\x1b[0m/g'

.PHONY: test test_pkg_db test_pkg_log bench_pkg_log test_pkg_web

test: test_pkg_db test_pkg_log test_pkg_web

test_pkg_db:
	go test -v ./pkg/db/... 2>&1 | $(COLORIZE)

test_pkg_web:
	go test -v ./pkg/web/... 2>&1 | $(COLORIZE)

test_pkg_logx:
	go test -v ./pkg/logx/... 2>&1 | $(COLORIZE)

test_pkg_mapper:
	go test -v ./pkg/mapper/... 2>&1 | $(COLORIZE)

bench_pkg_logx:
	go test -v ./pkg/logx/... -bench=. -benchmem 2>&1 | $(COLORIZE)

usage_basic:
	go run ./pkg/logx/examples/basic_usage.go | jq