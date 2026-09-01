package common

type ServiceGroup struct {
	AsyncService AsyncService
}

func NewServiceGroup() ServiceGroup {
	return ServiceGroup{
		AsyncService: NewAsyncService(),
	}
}
