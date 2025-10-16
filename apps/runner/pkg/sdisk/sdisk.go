package sdisk

func New(config Config) (DiskManager, error) {
	return NewManager(config)
}
