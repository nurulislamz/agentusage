package daemon

import (
	"context"
	"testing"
	"time"
)

func TestReadModelCacheIntervalRespectsPollInterval(t *testing.T) {
	tests := []struct {
		name string
		in   time.Duration
		want time.Duration
	}{
		{name: "default", in: 0, want: 30 * time.Second},
		{name: "minimum", in: time.Second, want: 5 * time.Second},
		{name: "normal", in: 30 * time.Second, want: 30 * time.Second},
		{name: "long", in: time.Hour, want: time.Hour},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := readModelCacheInterval(tt.in); got != tt.want {
				t.Fatalf("readModelCacheInterval(%s) = %s, want %s", tt.in, got, tt.want)
			}
		})
	}
}

func TestServiceContext(t *testing.T) {
	// 1. Nil service with nil fallback
	var nilSvc *Service
	if ctx := nilSvc.serviceContext(nil); ctx == nil {
		t.Error("serviceContext(nil) should return non-nil context")
	}

	// 2. Nil service with custom fallback
	customCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if ctx := nilSvc.serviceContext(customCtx); ctx != customCtx {
		t.Error("serviceContext(custom) should return fallback context")
	}

	// 3. Service with set ctx
	svcCtx, cancel2 := context.WithCancel(context.Background())
	defer cancel2()
	svc := &Service{ctx: svcCtx}
	if ctx := svc.serviceContext(customCtx); ctx != svcCtx {
		t.Error("serviceContext should prefer svc.ctx over fallback")
	}
}

func TestComputeReadModel_Empty(t *testing.T) {
	svc := &Service{}
	res, err := svc.computeReadModel(context.Background(), ReadModelRequest{
		Accounts: nil,
	})
	if err != nil {
		t.Fatalf("computeReadModel on empty accounts err: %v", err)
	}
	if len(res) != 0 {
		t.Errorf("computeReadModel = %+v, want empty map", res)
	}
}
