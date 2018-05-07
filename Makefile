VERSION				:=	$(shell cat ./VERSION)
ROOTFS_IMAGE		:=	cirocosta/xfsvol-rootfs
ROOTFS_CONTAINER	:=	rootfs
PLUGIN_NAME			:=	xfsvol
PLUGIN_FULL_NAME	:=	cirocosta/xfsvol


all: install


install:
	cd ./xfsvolctl && \
		go install \
			-ldflags "-X main.version=$(VERSION)" \
			-v
	cd ./main && \
		go install \
			-ldflags "-X main.version=$(VERSION)" \
			-v


test:
	go test ./... -v


fmt:
	go fmt ./...


rootfs-image:
	docker build -t $(ROOTFS_IMAGE) .


rootfs: rootfs-image
	docker rm -vf $(ROOTFS_CONTAINER) || true
	docker create --name $(ROOTFS_CONTAINER) $(ROOTFS_IMAGE) || true
	mkdir -p plugin/rootfs
	rm -rf plugin/rootfs/*
	docker export $(ROOTFS_CONTAINER) | tar -x -C plugin/rootfs
	docker rm -vf $(ROOTFS_CONTAINER)


plugin: rootfs
	docker plugin disable $(PLUGIN_NAME) || true
	docker plugin rm --force $(PLUGIN_NAME) || true
	docker plugin create $(PLUGIN_NAME) ./plugin
	docker plugin enable $(PLUGIN_NAME)


plugin-push: rootfs
	docker plugin rm --force $(PLUGIN_FULL_NAME) || true
	docker plugin create $(PLUGIN_FULL_NAME) ./plugin
	docker plugin create $(PLUGIN_FULL_NAME):$(VERSION) ./plugin
	docker plugin push $(PLUGIN_FULL_NAME)
	docker plugin push $(PLUGIN_FULL_NAME):$(VERSION)


.PHONY: install deps fmt rootfs-image rootfs plugin plugin-logs plugin-exec test
