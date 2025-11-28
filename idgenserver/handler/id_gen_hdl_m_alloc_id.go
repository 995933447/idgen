package handler

import (
	"context"

	"github.com/995933447/fastlog"
	"github.com/995933447/idgen/idgen"
	"github.com/995933447/idgen/idgenserver/idgenerator"
)

func (s *IdGen) MAllocId(ctx context.Context, req *idgen.MAllocIdReq) (*idgen.MAllocIdResp, error) {
	var resp idgen.MAllocIdResp

	ids, err := idgenerator.MAlloc(ctx, req.TbName, req.Count)
	if err != nil {
		fastlog.Error(err)
		return nil, err
	}

	resp.List = ids

	return &resp, nil
}
