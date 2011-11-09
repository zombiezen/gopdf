package pdf

// name is a PDF name object.
type name string

func (n name) String() string {
	return string(n)
}

// array is an ordered collection of PDF objects.
type array []interface{}

// dictionary is a mapping of PDF objects.
type dictionary map[name]interface{}

// stream is a blob of data.
type stream struct {
	Dictionary dictionary
	Bytes      []byte
}

type indirectObject struct {
	Number     uint
	Generation uint
	Object     interface{}
}
