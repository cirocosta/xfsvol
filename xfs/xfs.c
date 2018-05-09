#include "./xfs.h"

int
xfs_get_project_id(const char* dir)
{
	int            err = 0;
	int            temp_errno;
	int            dir_fd;
	struct fsxattr fs_xattr = { 0 };

	dir_fd = open(dir, O_DIRECTORY | O_PATH);
	if (dir_fd == -1) {
		return -1;
	}

	err = ioctl(dir_fd, FS_IOC_FSGETXATTR, &fs_xattr);
	if (err == -1) {
		temp_errno = errno;
		close(dir_fd);
		errno = temp_errno;
		return -1;
	}

	close(dir_fd);
	return fs_xattr.fsx_projid;
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
