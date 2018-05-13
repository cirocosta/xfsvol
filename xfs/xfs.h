#ifndef __XFS_H
#define __XFS_H

#ifndef _GNU_SOURCE
#define _GNU_SOURCE
#endif

#include <dirent.h>
#include <errno.h>
#include <fcntl.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>

#include <linux/fs.h>
#include <linux/quota.h>
#include <xfs/xqm.h>

#include <sys/ioctl.h>
#include <sys/quota.h>
#include <sys/stat.h>

/**
 * The size of a disk block as defined by the quota api.
 */
#define BASIC_BLOCK_SIZE 512

/**
 * The following definitions are mostly for compatibility
 * with older kernels.
 *
 * This is meant to make it work at least with 4.4+.
 *
 * For some definitions, check out the kernel tree.
 *
 * From linux/fs/xfs/libxfs/xfs_quota_defs.h
 * (https://github.com/torvalds/linux/blob/master/fs/xfs/libxfs/xfs_quota_defs.h#L37):
 *
 *      #define XFS_DQ_PROJ          0x0002
 *
 */

#ifndef FS_XFLAG_PROJINHERIT
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

/**
 * Provides the configuration to be used when
 * invoking the xfs getter and setter commands.
 */
typedef struct xfs_quota {
	__u64 size;
	__u64 inodes;
	__u64 used_size;
	__u64 used_inodes;
} xfs_quota_t;

/**
 * Sets the project quota for a given path as
 * specified in the arguments provided via the
 * argument `cfg`.
 *
 * Returns -1 in case of errors
 */
int
xfs_set_project_quota(const char*  fs_block_dev,
                      __u32        project_id,
                      xfs_quota_t* quota);

/**
 * Retrieves the quota configuration associated
 * with a given path as controled by a specified
 * backing fs block device.
 *
 * The values relative to the project quota are stored
 * in the provided `quota` reference passed as reference.
 *
 * Returns -1 in case of errors.
 */
int
xfs_get_project_quota(const char*  fs_block_dev,
                      __u32        project_id,
                      xfs_quota_t* quota);

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
 * `dir` represents the directory (which must already
 * exist) in which `filename` (name of the special file)
 * should be created.
 *
 * Returns -1 in case of errors.
 */
int
xfs_create_fs_block_dev(const char* dir, const char* filename);

#endif
