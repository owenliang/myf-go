package cron

// 封装Myf框架Job，为了可以利用WrappedJob功能
type MyfJob struct {
	f func()
}

func newJob(f func()) (job *MyfJob){
	job = &MyfJob{
		f: f,
	}
	return
}

func (myfJob *MyfJob) Run() {
	myfJob.f()
}
