package boot

import (
	"context"

	"github.com/995933447/fastlog"
	"github.com/995933447/idgen/idgenserver/idgenerator"
)

func InitApp() {
	if err := idgenerator.Init(context.TODO()); err != nil {
		fastlog.Fatalf("idgenerator.Init err:%v", err)
	}
}
