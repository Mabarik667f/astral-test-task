package doccmd

type GetDocumentsListCmd struct {
	Login *string
	Key   string
	Value any
	Limit int
}
