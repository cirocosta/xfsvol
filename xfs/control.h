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

#endif
