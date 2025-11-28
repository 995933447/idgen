package idgen

import (
	"context"

	"github.com/995933447/easymicro/grpc"
)

func PrepareGRPC(ctx context.Context, discoveryName string) error {
	if err := grpc.PrepareDiscoverGRPC(context.TODO(), EasymicroGRPCSchema, discoveryName); err != nil {
		return err
	}
	return nil
}
