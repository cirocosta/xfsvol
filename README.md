<h1 align="center">xfsvol 📂  </h1>

<h5 align="center">Docker Volume Plugin for managing local XFS-based volumes</h5>

<br/>

### Quickstart

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

### `xfsvolctl`

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

