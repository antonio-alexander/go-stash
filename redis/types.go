package redis

type Redis interface {
	//Configure
	Configure(...interface{}) error

	//SetParameters
	SetParameters(...interface{})

	//Initialize can be used to setup internal pointers
	// and ready the stash for usage
	Initialize() error

	//Shutdown can be used to tear down internal pointers
	// and ready the stash for garbage collection (or reuse)
	Shutdown() error
}
