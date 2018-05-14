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
  create_xfs_loopback_device_without_project_quota
  create_testing_directory

  lsblk
}

install_dependencies() {
  echo "INFO:
  Installing base dependencies using 'apt'.
  "

  sudo apt update -y
  sudo apt install -y \
    xfsprogs \
    xfslibs-dev \
    tree

  echo "SUCCESS:
  Dependencies installed.
  "
}

create_xfs_loopback_device() {
  echo "INFO:
  Creating XFS loopback device
  "

  sudo dd if=/dev/zero of=/xfs.512M.1 bs=1M count=512
  sudo losetup /dev/loop0 /xfs.512M.1
  sudo mkfs -t xfs /dev/loop0
  sudo mkdir -p /mnt/xfs
  sudo mount /dev/loop0 /mnt/xfs -o pquota

  echo "SUCCESS:
  Device created.
  "
}

create_xfs_loopback_device_without_project_quota() {
  echo "INFO:
  Creating XFS loopback device without project
  quota support.
  "

  sudo dd if=/dev/zero of=/xfs.512M.2 bs=1M count=512
  sudo losetup /dev/loop1 /xfs.512M.2
  sudo mkfs -t xfs /dev/loop1
  sudo mkdir -p /mnt/xfs-without-quota
  sudo mount /dev/loop1 /mnt/xfs-without-quota -o noquota

  echo "SUCCESS:
  Device created.
  "
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
