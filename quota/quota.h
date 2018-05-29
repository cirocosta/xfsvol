#ifndef __QUOTA_H
#define __QUOTA_H

#include <linux/types.h>

/**
 * Provides the configuration to be used when
 * invoking quota getter and setter commands.
 */
typedef struct quota {
	__u64 size;
	__u64 inodes;
	__u64 used_size;
	__u64 used_inodes;
} quota_t;

#endif
