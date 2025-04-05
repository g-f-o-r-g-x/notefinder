package types

type Writable bool
type Implementation interface {
	LoadData() (map[uint64]*Note, error)
	PutData(*Note) error
	UpdateData(*Note, *Note) error
	DeleteData(*Note) error
	SupportedProperties() map[string]Writable
	CanWrite() (bool, error)
}

type NotebookType int

const (
	NotebookConfigured NotebookType = iota
	NotebookAutoDiscovered
)

type Notebook struct {
	Name           string
	Config         map[string]string
	Data           map[uint64]*Note
	implementation Implementation
	Type           NotebookType
	Enabled        bool
}

func NewNotebook(name string, impl Implementation, config map[string]string,
	type_ NotebookType) *Notebook {
	return &Notebook{Name: name, implementation: impl, Config: config,
		Data: make(map[uint64]*Note), Type: type_, Enabled: true}
}

func (self *Notebook) LoadData() (map[uint64]*Note, error) {
	data, err := self.implementation.LoadData()
	return data, err
}

func (self *Notebook) CanWrite() (bool, error) {
	return self.implementation.CanWrite()
}

func (self *Notebook) PutData(note *Note) error {
	return self.implementation.PutData(note)
}

func (self *Notebook) UpdateData(oldNote *Note, newNote *Note) error {
	return self.implementation.UpdateData(oldNote, newNote)
}

func (self *Notebook) DeleteData(note *Note) error {
	return self.implementation.DeleteData(note)
}
