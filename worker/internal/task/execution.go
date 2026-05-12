package task

type ExecSpec struct {
	Command string
	Image   string
	WorkDir string
}

type Execution struct {
	ExitCode int
}
