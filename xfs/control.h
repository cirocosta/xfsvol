#ifndef __CONTROL_H
#define __CONTROL_H

#include <dirent.h>
#include <linux/dqblk_xfs.h>
#include <linux/fs.h>
#include <linux/quota.h>
#include <stdlib.h>

#ifndef FS_XFLAG_PROJINHERIT

struct fsxattr {
	__u32         fsx_xflags;
	__u32         fsx_extsize;
	__u32         fsx_nextents;
	__u32         fsx_projid;
	unsigned char fsx_pad[12];
};

#define FS_XFLAG_PROJINHERIT 0x00000200
#endif

#ifndef FS_IOC_FSGETXATTR
#define FS_IOC_FSGETXATTR _IOR('X', 31, struct fsxattr)
#endif

#ifndef FS_IOC_FSSETXATTR
#define FS_IOC_FSSETXATTR _IOW('X', 32, struct fsxattr)
#endif

#ifndef PRJQUOTA
#define PRJQUOTA 2
#endif

#ifndef XFS_PROJ_QUOTA
#define XFS_PROJ_QUOTA 2
#endif

#ifndef Q_XSETPQLIM
#define Q_XSETPQLIM QCMD(Q_XSETQLIM, PRJQUOTA)
#endif

#ifndef Q_XGETPQUOTA
#define Q_XGETPQUOTA QCMD(Q_XGETQUOTA, PRJQUOTA)
#endif

/**
 * Provides the configuration to be used when
 * invoking the xfs getter and setter commands.
 */
typedef struct xfs_quota {
	__u64 size;
	__u64 inodes;
} xfs_quota_t;

/**
 * Sets the project quota for a given path as
 * specified in the arguments provided via the
 * argument `cfg`.
 *
 * Returns -1 in case of errors
 */
int
xfs_set_project_quota(const char* fs_block_dev, __u32 project_id, xfs_quota_t*);

/**
 * Retrieves the quota configuration associated
 * with a given path as controled by a specified
 * backing fs block device.
 *
 * Returns NULL in case of errors.
 */
xfs_quota_t*
xfs_get_project_quota(const char* fs_block_dev, __u32 project_id);

/**
 * Sets the project_id of a given directory.
 *
 * Returns -1 in case of errors.
 */
int
xfs_set_project_id(const char* dir, __u32 project_id);

/**
 * Retrieves the project_id of a given targetPath
 * as managed by a backing_fs_block_dev.
 *
 * Returns -1 in case of errors.
 */
int
xfs_get_project_id(const char* dir);

/**
 * Creates the filesystem block device to control
 * xfs quotas under a given root.
 *
 * `filename` represents the absolute path to the
 * device to be created.
 *
 * Returns -1 in case of errors.
 */
int
xfs_create_fs_block_dev(const char* filename);

#endif
