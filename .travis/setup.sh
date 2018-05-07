#!/bin/bash

# Prepares the travis-ci machine to have the
# necessary components for properly testing
# the core xfsvol library.
#
# It makes sures that we have:
# - xfs utilities installed
# - testing directories set up
# - a xfs filesystem mounted in a loopback
#   device.

set -o errexit

main() {
	install_dependencies
	create_xfs_loopback_device
	create_testing_directory
}

install_dependencies() {
  echo "INFO:
  Installing base dependencies using 'apt'.
  "

	sudo apt update -y
	sudo apt install -y xfsprogs tree

	echo "SUCCESS:
  Dependencies installed.
  "
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

create_testing_directory() {
	echo "INFO:
  Creating testing directory '/mnt/xfs/tmp'.
  "

	sudo mkdir -p /mnt/xfs/tmp
	sudo chown -R $(whoami) /mnt/xfs/tmp

	echo "SUCCESS:
  Testing directory created
  "

	tree /mnt
}

main
