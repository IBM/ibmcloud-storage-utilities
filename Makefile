
.PHONY: block-storage-attacher/deps
block-storage-attacher/deps:
	cd block-storage-attacher; \
	make deps

.PHONY: block-storage-attacher
block-storage-attacher:
	cd block-storage-attacher; \
	make vet; \
	make fmt; \
	make test; \
	make lint; \
	cd ../ && pwd && make lint-root-repo; \
	cd block-storage-attacher && make coverage

.PHONY: lint-root-repo
 lint-root-repo:
	bt lint copyright
	bt lint shellcheck
