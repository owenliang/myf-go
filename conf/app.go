package conf

const (
	DEFAULT_READ_TIMEOUT   int = 5000
	DEFAULT_WRITE_TIMEOUT  int = 5000
	DEFAULT_HANDLE_TIMEOUT int = 5000
)

// APP配置
type AppConfig struct {
	Addr             string            `toml:"listen"`        // 监听地址
	ReadTimeout      int               `toml:"readTimeout"`   // 请求读超时
	HandleTimeout    int               `toml:"handleTimeout"` // 请求处理超时
	WriteTimeout     int               `toml:"writeTimeout"`  // 请求写超时
}

// 默认App配置
func DefaultAppConfig() (appConfig *AppConfig) {
	appConfig = &AppConfig{
		ReadTimeout: DEFAULT_READ_TIMEOUT,
		HandleTimeout: DEFAULT_HANDLE_TIMEOUT,
		WriteTimeout: DEFAULT_WRITE_TIMEOUT,
	}
	return
}
