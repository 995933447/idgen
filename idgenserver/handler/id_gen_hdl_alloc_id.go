package handler

import (
	"context"
	"errors"

	"github.com/995933447/fastlog"
	"github.com/995933447/idgen/idgen"
	"github.com/995933447/idgen/idgenserver/idgenerator"
)

func (s *IdGen) AllocId(ctx context.Context, req *idgen.AllocIdReq) (*idgen.AllocIdResp, error) {
	var resp idgen.AllocIdResp

	ids, err := idgenerator.MAlloc(ctx, req.TbName, 1)
	if err != nil {
		fastlog.Error(err)
		return nil, err
	}

	if len(ids) == 0 {
		return nil, errors.New("unknown error")
	}

	resp.Id = ids[0]

	return &resp, nil
}
