package dotfile

type Status int

const (
	Ok Status = iota
	Skipped
	Failed
)

func (s Status) String() string {
	switch s {
	case Ok:
		return "ok"
	case Skipped:
		return "skipped, already installed"
	case Failed:
		return "failed"
	default:
		return "unknown"
	}
}

type OperationStatus struct {
	Dotfile *Item
	Status  Status
	Error   error
}
