#!/bin/bash

set -o errexit

main() {
	install_dependencies
	create_xfs_loopback_device
	create_testing_directory
}

install_dependencies() {
	sudo apt update -y
	sudo apt install -y xfsprogs tree
}

create_testing_directory() {
	echo "INFO:
  Creating testing directory /mnt/xfs/tmp
  "

	sudo mkdir -p /mnt/xfs/tmp
	sudo chown -R $(whoami) /mnt/xfs/tmp

	echo "SUCCESS:
  Testing directory created
  "

	tree /mnt
}

create_xfs_loopback_device() {
	echo "INFO:
  Creating XFS loopback device
  "

	sudo dd if=/dev/zero of=/xfs.1G bs=1M count=1024
	sudo losetup /dev/loop0 /xfs.1G
	sudo mkfs -t xfs /dev/loop0
	sudo mkdir -p /mnt/xfs
	sudo mount /dev/loop0 /mnt/xfs -o pquota

	echo "SUCCESS:
  Device created.
  "

	lsblk
}

main
