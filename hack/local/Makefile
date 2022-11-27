.PHONY: run-k3d-with-eventing
run-k3d-with-eventing:
	@make -C ./eventing run

.PHONY: busola-enable
busola-enable:
	@make -C ./busola busola-enable

.PHONY: verify
verify:
	@make -C ./eventing verify

.PHONY: run
run: run-k3d-with-eventing busola-enable

.PHONY: test
test: run verify
	@echo '### eventing operator operational'

.PHONY: stop
stop:
	@make -C ./eventing stop
