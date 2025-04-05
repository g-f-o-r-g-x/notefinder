package types

type NoteKey struct {
	Notebook *Notebook
	UUID     uint64
}

type QueryMethod int

const (
	QueryWithIndex QueryMethod = iota
	QueryDirect
	QueryRegexp
	QueryComplex
)

type Query struct {
	Needle   string
	Haystack *Notebook
	Method   QueryMethod
}
