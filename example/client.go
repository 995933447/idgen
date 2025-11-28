package main

import (
	"context"
	"fmt"

	"github.com/995933447/easymicro/loader"
	"github.com/995933447/idgen/idgen"
)

func main() {
	if err := loader.LoadEtcdFromLocal(); err != nil {
		panic(err)
	}

	if err := loader.LoadDiscoveryFromLocal(); err != nil {
		panic(err)
	}

	if err := idgen.PrepareGRPC(context.TODO(), "default"); err != nil {
		panic(err)
	}

	allocResp, err := idgen.IdGenGRPC().AllocId(context.TODO(), &idgen.AllocIdReq{
		TbName: "user",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(allocResp.Id)

	mAllocResp, err := idgen.IdGenGRPC().MAllocId(context.TODO(), &idgen.MAllocIdReq{
		TbName: "user",
		Count:  10,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(mAllocResp.List)
}
