.PHONY: run-k3d-with-eventing
run-k3d-with-eventing:
	@make -C ./eventing run

.PHONY: verify
verify:
	@make -C ./eventing verify

.PHONY: run
run: run-k3d-with-eventing

.PHONY: test
test: run verify
	@echo '### eventing operator operational'

.PHONY: stop
stop:
	@make -C ./eventing stop
