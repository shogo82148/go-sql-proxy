package proxy

import (
	"database/sql/driver"
	"io"
)

var _ io.Closer = (*Connector)(nil)
var _ driver.Connector = (*Connector)(nil)
