<h1 align="center">xfsvol 📂  </h1>

<h5 align="center">Docker Volume Plugin for managing local XFS-based volumes</h5>

<br/>

## Quickstart

1. Create a mountpoint at `/mnt/xfs` and a directory `/mnt/xfs/volumes`. 

For testing purposes this mountpoint can be a loopback device (note: use a loopback device for testing purposes only).

```sh
sudo dd if=/dev/zero of=/xfs.1G bs=1M count=1024
sudo losetup /dev/loop0 /xfs.1G
sudo mkfs -t xfs -n ftype=1 /dev/loop0
sudo mkdir -p /mnt/xfs/volumes
sudo mount /dev/loop0 /mnt/xfs -o pquota
```

2. Install the plugin

```
docker plugin install \
        --grant-all-permissions \
        --alias xfsvol \
        cirocosta/xfsvol

docker plugin ls
ID                  NAME                DESCRIPTION                                   ENABLED
06545b643c6a        xfsvol:latest       Docker plugin to manage XFS-mounted volumes   true
```

3. Create a named volume

```
docker volume create \
        --driver xfsvol \
        --opt size=10M \
        myvolume1
```

4. Run a container with the volume attached

```
docker run -it \
        -v myvolume1:/myvolume1 \
        alpine /bin/sh

dd if=/dev/zero of=/myvolume1/file bs=1M count=100
(fail!)
```

5. Check the volumes list

```
docker volume ls
DRIVER              VOLUME NAME
xfsvol:latest       myvolume1 (1.004MB)
local               dockerdev-go-pkg-cache-gopath
local               dockerdev-go-pkg-cache-goroot-linux_amd64
local               dockerdev-go-pkg-cache-goroot-linux_amd64_netgo
```

and the `xfsvolctl` utility:

```
sudo /usr/bin/xfsvolctl ls --root /mnt/xfs/volumes/
NAME   QUOTA
ciro   1.004 MB
```

## `xfsvolctl`

This tool is made to help inspect the project quotas created under a given root path as well as create/delete others. It's usage is documented under `--help`:

```
xfsvolctl --help
NAME:
   xfsvolctl - Controls the 'xfsvol' volume plugin

USAGE:
   xfsvolctl [global options] command [command options] [arguments...]

VERSION:
   0.0.0

COMMANDS:
     ls       Lists the volumes managed by 'xfsvol' plugin
     create   Creates a volume with XFS project quota enforcement
     delete   Deletes a volume managed by 'xfsvol' plugin
     help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help
   --version, -v  print the version
```

### Under the hood


```
$ sudo mkdir -p /mnt/xfs/tmp/bbb
$ sudo strace -f xfsvolctl create \
        --root /mnt/xfs/tmp/bbb \
        --name ccc \
        --size 1024 \
        --inode 1024

stat("/mnt/xfs/tmp/bbb", {st_mode=S_IFDIR|0755, st_size=6, ...}) = 0
unlinkat(AT_FDCWD, "/mnt/xfs/tmp/bbb/backingFsBlockDev", 0) = -1 ENOENT (No such file or directory)
mknodat(AT_FDCWD, "/mnt/xfs/tmp/bbb/backingFsBlockDev", S_IFBLK|0600, makedev(7, 0)) = 0
quotactl(Q_XSETQLIM|PRJQUOTA, "/mnt/xfs/tmp/bbb/backingFsBlockDev", 1, {version=1, flags=XFS_PROJ_QUOTA, fieldmask=0xc, id=1, blk_hardlimit=0, blk_softlimit=0, ino_hardlimit=0, ino_softlimit=0, bcount=0, icount=0, ...}) = 0
...
mkdirat(AT_FDCWD, "/mnt/xfs/tmp/bbb/ccc", 0755) = 0
open("/mnt/xfs/tmp/bbb/ccc", O_RDONLY|O_NONBLOCK|O_DIRECTORY|O_CLOEXEC) = 3
fstat(3, {st_mode=S_IFDIR|0755, st_size=6, ...}) = 0
ioctl(3, FS_IOC_FSGETXATTR, 0xc4201a95e4) = 0
ioctl(3, FS_IOC_FSSETXATTR, 0xc4201a95e4) = 0

quotactl(Q_XSETQLIM|PRJQUOTA, "/mnt/xfs/tmp/bbb/backingFsBlockDev", 2, {version=1, flags=XFS_PROJ_QUOTA, fieldmask=0xc, id=2, blk_hardlimit=2, blk_softlimit=2, ino_hardlimit=1024, ino_softlimit=1024, bcount=0, icount=0, ...}) = 0
[pid  6833] ioctl(2, TCGETS, {B38400 opost isig icanon echo ...}) = 0


// retrieving (ls)

[pid  8609] quotactl(Q_XGETQUOTA|PRJQUOTA, "/mnt/xfs/tmp/ddd/backingFsBlockDev", 2, {version=1, flags=XFS_PROJ_QUOTA, fieldmask=0, id=2, blk_hardlimit=8, blk_softlimit=8, ino_hardlimit=0, ino_softlimit=0, bcount=8, icount=104, ...}) = 0


```

