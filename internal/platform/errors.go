package platform

type BluetoothErrorCategory int

const (
	BluetoothErrorUnavailable BluetoothErrorCategory = iota
	BluetoothErrorPermissionDenied
	BluetoothErrorDisabled
	BluetoothErrorUnknown
)

type BluetoothError struct {
	Category BluetoothErrorCategory
	Message  string
}

func (e *BluetoothError) Error() string {
	if e == nil {
		return ""
	}
	return e.Message
}
