# plugin

A plugin consists mostly of an [opencontainers/runc](https://github.com/opencontainers/runc) container definition. That is, the set `{rootfs/, config.json}`.
To generate the `rootfs` directory head to the root of this project and run `make rootfs`. 

Note.: `make rootfs` requires `docker`.

