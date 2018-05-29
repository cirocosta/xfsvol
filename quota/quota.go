package quota

// Quota defines the limit params to be applied or that
// are already set to a project.
//
// Fields named `Used` are meant to `Get` operations only,
// to display how much of the quotah as been used so far.
type Quota struct {
	// Size represents total size that can be commited
	// to a tree of directories under this quota.
	Size uint64

	// INode tells the maximum number of INodes that can be created;
	INode uint64

	// UsedSize is the disk size that has been used under quota;
	UsedSize uint64

	// UsedINode is the number of INodes that used so far.
	UsedInode uint64
}
