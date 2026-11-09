SHELL := /bin/sh

.PHONY: help check test fmt
help:
	@echo '기초: make -C go-basic basics | concurrency | check'
	@echo '서버: make -C go-practice up | down | db | migrate | test-db'
	@echo '공통: make check | test | fmt'
check:
	$(MAKE) -C go-basic check
	$(MAKE) -C go-practice check
test:
	$(MAKE) -C go-basic test
	$(MAKE) -C go-practice test
fmt:
	$(MAKE) -C go-basic fmt
	$(MAKE) -C go-practice fmt
