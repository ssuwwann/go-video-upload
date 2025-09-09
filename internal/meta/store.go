package meta

// Store abstracts metadata persistence.
type Store interface {
	Create(m Metadata) error
	Get(id string) (Metadata, error)
	Update(m Metadata) error
	List() ([]Metadata, error)
}
