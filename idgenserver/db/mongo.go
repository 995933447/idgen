package db

import (
	"github.com/995933447/idgen/idgen"
	"github.com/995933447/idgen/idgenserver/config"
)

func NewIdSegmentModel() *idgen.IdSegmentModel {
	mod := idgen.NewIdSegmentModel()
	config.SafeReadServerConfig(func(c *config.ServerConfig) {
		mod.SetConn(c.GetMongoConn())
		mod.SetDb(c.GetMongoDb())
	})
	return mod
}
