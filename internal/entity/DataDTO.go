package entity

type DataDTO struct {
	Data     []byte   `json:"data"`
	DataType DataType `json:"data_type"`
}
