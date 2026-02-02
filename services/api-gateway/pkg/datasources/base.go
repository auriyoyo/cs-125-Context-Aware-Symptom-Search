package datasources

import "context"

type DataSource interface {
	Name() string
	DatabaseName() string
	Query(query string) error
	Start(ctx context.Context) error
	Stop() error
}





