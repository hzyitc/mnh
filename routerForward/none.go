package routerForward

type None struct {
}

func NewTCPForward_None(port int) (Interface, error) {
	return &None{}, nil
}

func NewUDPForward_None(port int) (Interface, error) {
	return &None{}, nil
}

func (u *None) Redo() error {
	return nil
}

func (u *None) Close() {
}
