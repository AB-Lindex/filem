package storage

// Storage is the abstract api to access a generic `storage`
type Storage interface {
	Save(name, location, ct string, size int64, dryRun bool) (*StoredObject, error)
}

// StoredObject is the datamodel for the resulting message
type StoredObject struct {
	URL     string            `yaml:"url" json:"url"`
	Headers map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
}
