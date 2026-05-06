package images

type ShutdownFunc func()

func InitRuntime() (ShutdownFunc, error) {
	return initRuntime()
}

