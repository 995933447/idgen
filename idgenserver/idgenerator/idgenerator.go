package idgenerator

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/995933447/fastlog"
	"github.com/995933447/idgen/idgen"
	"github.com/995933447/idgen/idgenserver/config"
	"github.com/995933447/idgen/idgenserver/db"
	"github.com/995933447/mgorm"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	generator *Generator
	mu        sync.RWMutex
)

type Generator struct {
	mu       sync.RWMutex
	allocMap map[string]*tbIdAlloc
}

func (g *Generator) getAlloc(ctx context.Context, tbName string) (*tbIdAlloc, error) {
	g.mu.RLock()
	alloc, ok := g.allocMap[tbName]
	g.mu.RUnlock()
	if !ok {
		g.mu.Lock()
		defer g.mu.Unlock()

		if alloc, ok = g.allocMap[tbName]; !ok {
			alloc = newTbAlloc(tbName)
			if err := alloc.allocIdSegment(ctx); err != nil {
				return nil, err
			}
			g.allocMap[tbName] = alloc
		}
	}
	return alloc, nil
}

type tbId struct {
	value uint64
	onUse func()
}

func newTbAlloc(tbName string) *tbIdAlloc {
	return &tbIdAlloc{
		tbName: tbName,
		exitCh: make(chan struct{}, 1),
	}
}

type tbIdAlloc struct {
	tbName  string
	idCh    chan *tbId
	exitCh  chan struct{}
	stopped bool
}

func (a *tbIdAlloc) fetchId() uint64 {
	id := <-a.idCh
	if id.onUse != nil {
		id.onUse()
	}
	return id.value
}

func (a *tbIdAlloc) exit() {
	a.exitCh <- struct{}{}
	a.stopped = true
}

// 分配下一个步长区间的id到缓存中
func (a *tbIdAlloc) allocIdSegment(ctx context.Context) error {
	mod := db.NewIdSegmentModel()

	segment, err := mod.FindOneByTbName(ctx, a.tbName)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return err
		}

		config.SafeReadServerConfig(func(c *config.ServerConfig) {
			segment = &idgen.IdSegmentOrm{
				TbName: a.tbName,
				MaxId:  c.GetStep(),
				Step:   c.GetStep(),
			}
		})

		err = db.NewIdSegmentModel().InsertOne(ctx, segment)
		if err != nil {
			if !mgorm.IsUniqIdxConflictError(err) {
				fastlog.Error(err)
				return err
			}
		} else {
			a.fillAllocIds(1, segment.MaxId)
			return nil
		}

		segment, err = mod.FindOneByTbName(ctx, a.tbName)
		if err != nil {
			fastlog.Error(err)
			return err
		}
	}

	for {
		var step uint64
		config.SafeReadServerConfig(func(c *config.ServerConfig) {
			step = c.GetStep()
		})

		nextMaxId := segment.MaxId + step
		res, err := mod.UpdateOne(ctx, bson.M{
			"tb_name": a.tbName,
			"max_id":  segment.MaxId,
		}, bson.M{
			"max_id": nextMaxId,
			"step":   step,
		})
		if err != nil {
			fastlog.Error(err)
			return err
		}

		if res.ModifiedCount == 0 {
			segment, err = mod.FindOneByTbName(ctx, a.tbName)
			if err != nil {
				fastlog.Error(err)
				return err
			}

			continue
		}

		a.fillAllocIds(segment.MaxId+1, nextMaxId)

		break
	}

	return nil
}

// 填充区间id到缓存队列中,等待被使用
func (a *tbIdAlloc) fillAllocIds(startId, maxId uint64) {
	if a.idCh == nil {
		a.idCh = make(chan *tbId, maxId)
	}

	var setAllocNextIdSegmentCallback bool
	var ids []*tbId
	for i := startId; i <= maxId; i++ {
		var onUse func() = nil
		// 当缓存中id队列不足50个或者不足约10%,提取分配下个步长区间的id
		if !setAllocNextIdSegmentCallback && (maxId-i <= 50 || float64(i) > float64(maxId)*0.9) {
			setAllocNextIdSegmentCallback = true
			onUse = func() {
				a.stillTryAllocIdSegment()
			}
		}

		ids = append(ids, &tbId{
			value: i,
			onUse: onUse,
		})
	}

	go func() {
		for _, id := range ids {
			select {
			case a.idCh <- id:
			case <-a.exitCh:
				select {
				case a.exitCh <- struct{}{}:
				}
				return
			}
		}
	}()
}

func (a *tbIdAlloc) stillTryAllocIdSegment() {
	go func() {
		var failedCount int
		for {
			// 已经停止了
			if a.stopped {
				return
			}

			err := a.allocIdSegment(context.TODO())
			if err != nil {
				fastlog.Errorf("allocIdSegment err: %v", err)
			} else {
				break
			}

			failedCount++

			if failedCount > 5 {
				fastlog.Errorf("allocIdSegment failed count reach %d", failedCount)
				time.Sleep(time.Second * 5)
			}

			continue
		}
	}()
}

func Init(ctx context.Context) error {
	if generator != nil {
		return nil
	}

	mu.Lock()
	defer mu.Unlock()

	if generator != nil {
		return nil
	}

	segments, err := db.NewIdSegmentModel().FindAll(ctx, bson.M{}, bson.M{"tb_name": 1})
	if err != nil {
		return err
	}

	g := &Generator{
		allocMap: make(map[string]*tbIdAlloc),
	}

	for _, segment := range segments {
		alloc := newTbAlloc(segment.TbName)

		if err = alloc.allocIdSegment(ctx); err != nil {
			for _, a := range g.allocMap {
				a.exit()
			}
			return err
		}

		g.allocMap[segment.TbName] = alloc
	}

	generator = g

	return nil
}

func getGenerator() (*Generator, error) {
	if generator == nil {
		return nil, errors.New("generator not initialized")
	}
	return generator, nil
}

func MAlloc(ctx context.Context, tbName string, n int32) ([]uint64, error) {
	g, err := getGenerator()
	if err != nil {
		return nil, err
	}

	alloc, err := g.getAlloc(ctx, tbName)
	if err != nil {
		return nil, err
	}

	var ids []uint64
	for i := int32(0); i < n; i++ {
		ids = append(ids, alloc.fetchId())
	}

	return ids, nil
}
