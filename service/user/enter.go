package user

type ServiceGroup struct {
	IndexService IndexService
}

func NewServiceGroup() ServiceGroup {
	return ServiceGroup{
		IndexService: NewIndexService(),
	}
}
