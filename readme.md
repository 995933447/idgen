## 基于leaf segment算法+内存缓存+mongodb持久化实现高性能分布式id生成器服务

使用例子:
````go
package main

import (
	"context"
	"fmt"

	"github.com/995933447/easymicro/loader"
	"github.com/995933447/idgen/idgen"
)

func main() {
	// 加载etcd
	if err := loader.LoadEtcdFromLocal(); err != nil {
		panic(err)
	}

	// 初始化服务注册发现组件
	if err := loader.LoadDiscoveryFromLocal(); err != nil {
		panic(err)
	}

	// 为grpc初始化
	if err := idgen.PrepareGRPC(context.TODO(), "default"); err != nil {
		panic(err)
	}

	// rpc调用生成一个id
	allocResp, err := idgen.IdGenGRPC().AllocId(context.TODO(), &idgen.AllocIdReq{
		TbName: "user",
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(allocResp.Id) 
	// 输出：38001

	// rpc调用生成10个id
	mAllocResp, err := idgen.IdGenGRPC().MAllocId(context.TODO(), &idgen.MAllocIdReq{
		TbName: "user",
		Count:  10,
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(mAllocResp.List)
	// 输出：[38002 38003 38004 38005 38006 38007 38008 38009 38010 38011]
}

````