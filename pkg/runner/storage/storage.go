package storage

type Plugins interface {
	LoadScriptByName(string) (string, error)
}

type Storage interface {
	Plugins() Plugins
}
