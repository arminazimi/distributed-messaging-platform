package backpressure

import (
	"context"
	"testing"
)

type fakeGetter struct {
	count int64
	err   error
}

func (f fakeGetter) GetContext(_ context.Context, dest any, _ string, _ ...any) error {
	if f.err != nil {
		return f.err
	}
	*(dest.(*int64)) = f.count
	return nil
}

func TestCheckPassesUnderThreshold(t *testing.T) {
	result, err := Check(context.Background(), fakeGetter{count: 9}, true, 10)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if result.Active {
		t.Fatalf("expected backpressure to be inactive")
	}
}

func TestCheckRejectsOverThreshold(t *testing.T) {
	result, err := Check(context.Background(), fakeGetter{count: 11}, true, 10)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if !result.Active {
		t.Fatalf("expected backpressure to be active")
	}
}

func TestCheckDisabledAllowsRequests(t *testing.T) {
	result, err := Check(context.Background(), fakeGetter{count: 1000}, false, 10)
	if err != nil {
		t.Fatalf("check: %v", err)
	}
	if result.Active {
		t.Fatalf("expected disabled backpressure to allow requests")
	}
}
