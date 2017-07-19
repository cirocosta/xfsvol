ROOTFS_IMAGE		:= cirocosta/xfsvol-rootfs
ROOTFS_CONTAINER	:= rootfs
PLUGIN_NAME			:= xfsvol
PLUGIN_FULL_NAME	:= cirocosta/xfsvol
PLUGIN_ID			:= $(shell docker plugin inspect $(PLUGIN_NAME) --format '{{ .ID }}')


all: build


build:
	cd ./xfsvolctl && go build -v


install:
	cd ./xfsvolctl && go install -v
	cd ./main && go install -v


test:
	cd ./manager && go test -v
	cd ./lib && go test -v


fmt:
	cd ./manager && go fmt
	cd ./lib && go fmt
	cd ./main && go fmt
	cd ./xfsvolctl && go fmt


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
	docker plugin push $(PLUGIN_FULL_NAME)


plugin-logs:
	docker run \
		--rm \
		-it \
		--privileged \
		--pid=host \
		justincormack/nsenter1 \
		/bin/sh -c 'docker-runc exec $(PLUGIN_ID) tail -n 100 -f /var/log/xfsvol/plugin.log'


plugin-exec:
	docker run \
		--rm \
		-it \
		--privileged \
		--pid=host \
		justincormack/nsenter1 \
		/bin/sh -c 'docker-runc exec -t $(PLUGIN_ID) sh'


.PHONY: install deps fmt rootfs-image rootfs plugin plugin-logs plugin-exec test
