package conf

const DEFAULT_CRON_WAIT = 5000  // 默认退出等待时间

// Cron配置
type CronConfig struct {
	WaitGraceExit int `toml:"waitGraceExit"`
}

// 默认Cron配置
func DefaultCronConfig() (cronConfig *CronConfig) {
	cronConfig = &CronConfig{
		DEFAULT_CRON_WAIT,
	}
	return
}