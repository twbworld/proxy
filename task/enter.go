package task

type Manager struct{}

// 任务管理器
func NewManager() *Manager {
	return new(Manager)
}
