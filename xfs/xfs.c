#include "./xfs.h"

int
xfs_set_project_quota(const char*  fs_block_dev,
                      __u32        project_id,
                      xfs_quota_t* quota)
{
	int             err        = 0;
	fs_disk_quota_t disk_quota = {
		.d_version       = FS_DQUOT_VERSION,
		.d_id            = project_id,
		.d_flags         = XFS_PROJ_QUOTA,
		.d_blk_hardlimit = quota->size / 512,
		.d_blk_softlimit = quota->size / 512,
		.d_ino_hardlimit = quota->inodes,
		.d_ino_softlimit = quota->inodes,
		.d_fieldmask =
		  FS_DQ_BHARD | FS_DQ_BSOFT | FS_DQ_ISOFT | FS_DQ_IHARD,
	};

	err = quotactl(QCMD(Q_XSETQLIM, PRJQUOTA),
	               fs_block_dev,
	               project_id,
	               (void*)&disk_quota);
	if (err == -1) {
		return -1;
	}

	return 0;
}

int
xfs_get_project_quota(const char*  fs_block_dev,
                      __u32        project_id,
                      xfs_quota_t* quota)
{
	int             err        = 0;
	fs_disk_quota_t disk_quota = { 0 };

	err = quotactl(QCMD(Q_XGETQUOTA, PRJQUOTA),
	               fs_block_dev,
	               project_id,
	               (void*)&disk_quota);
	if (err == -1) {
		return -1;
	}

	quota->size   = disk_quota.d_blk_hardlimit * 512;
	quota->inodes = disk_quota.d_ino_hardlimit;

	return 0;
}

int
xfs_get_project_id(const char* dir)
{
	int            err = 0;
	int            save_errno;
	int            dir_fd;
	struct fsxattr fs_xattr = { 0 };

	dir_fd = open(dir, O_RDONLY | O_DIRECTORY);
	if (dir_fd == -1) {
		return -1;
	}

	err = ioctl(dir_fd, FS_IOC_FSGETXATTR, &fs_xattr);
	if (err == -1) {
		save_errno = errno;
		close(dir_fd);
		errno = save_errno;
		return -1;
	}

	close(dir_fd);
	return fs_xattr.fsx_projid;
}

int
xfs_set_project_id(const char* dir, __u32 project_id)
{
	int            dir_fd;
	int            err = 0;
	int            save_errno;
	struct fsxattr fs_xattr = { 0 };

	dir_fd = open(dir, O_RDONLY | O_DIRECTORY);
	if (dir_fd == -1) {
		return -1;
	}

	err = ioctl(dir_fd, FS_IOC_FSGETXATTR, &fs_xattr);
	if (err == -1) {
		save_errno = errno;
		close(dir_fd);
		errno = save_errno;
		return -1;
	}

	fs_xattr.fsx_projid = project_id;
	fs_xattr.fsx_xflags |= FS_XFLAG_PROJINHERIT;

	err = ioctl(dir_fd, FS_IOC_FSSETXATTR, &fs_xattr);
	if (err == -1) {
		save_errno = errno;
		close(dir_fd);
		errno = save_errno;
		return -1;
	}

	close(dir_fd);
	return 0;
}

int
xfs_create_fs_block_dev(const char* dir, const char* filename)
{
	int         err      = 0;
	struct stat stat_buf = { 0 };
	char        full_path[PATH_MAX];

	if (strlen(dir) + strlen(filename) + 2 > PATH_MAX) {
		err   = -1;
		errno = ENAMETOOLONG;
		return err;
	}

	err = stat(dir, &stat_buf);
	if (err == -1) {
		return err;
	}

	if ((stat_buf.st_mode & S_IFMT) != S_IFDIR) {
		err   = -1;
		errno = ENOTDIR;
		return err;
	}

	strcpy(full_path, dir);
	strcat(full_path, "/");
	strcat(full_path, filename);

	err = mknod(full_path, S_IFBLK | 0600, stat_buf.st_dev);
	if (err == -1) {
		return err;
	}

	return err;
}
