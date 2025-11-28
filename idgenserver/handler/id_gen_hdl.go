package handler

import (
	"github.com/995933447/idgen/idgen"
)

type IdGen struct {
	idgen.UnimplementedIdGenServer
	ServiceName string
}

var IdGenHandler = &IdGen{
	ServiceName: idgen.EasymicroGRPCPbServiceNameIdGen,
}