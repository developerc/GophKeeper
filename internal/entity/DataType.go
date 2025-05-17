package entity

import "fmt"

type DataType uint8

const (
	RAW DataType = iota + 1
	CRED
	FILE
	CARD
)

func (dt DataType) String() string {
	switch dt {
	case RAW:
		return "RAW"
	case CRED:
		return "CRED"
	case FILE:
		return "FILE"
	case CARD:
		return "CARD"
	default:
		return fmt.Sprintf("%d", int(dt))
	}
}
