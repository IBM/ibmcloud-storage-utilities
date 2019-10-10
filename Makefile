
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
        make gosec; \
	make coverage

