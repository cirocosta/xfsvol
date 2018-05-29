package quota

type Control interface {
	GetQuota(targetPath string) (q *Quota, err error)
	SetQuota(targetPath string, quota Quota) (err error)
}
