package project

// Task is a falcula task. It is built over the falcula Script type
type Task struct {
	Script `yaml:",inline"`
}
