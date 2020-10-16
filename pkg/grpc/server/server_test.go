package server

import (
	"context"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/object88/tugboat/mocks"
	"github.com/object88/tugboat/mocks/grpc/sample"
	"github.com/object88/tugboat/pkg/logging/testlogger"
	"google.golang.org/grpc"
)

func Test_GRPC_Server_NoRegisterFuncs(t *testing.T) {
	s, err := New(testlogger.TestLogger{T: t}, 4000)
	if s != nil {
		t.Fatalf("Did not expect to get `serve` back")
	}
	if err == nil {
		t.Errorf("Did not get eexpected error")
	}
}

func Test_GRPC_Server(t *testing.T) {
	makeHandler := func(ctrl *gomock.Controller) *mocks.MockHandler { return mocks.NewMockHandler(ctrl) }
	tcs := []struct {
		name         string
		handlerCount int
	}{
		{
			name:         "one-handler",
			handlerCount: 1,
		},
		{
			name:         "three-handlers",
			handlerCount: 3,
		},
	}
	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			tl := testlogger.TestLogger{T: t}

			handlers := []Handler{}
			for i := 0; i < tc.handlerCount; i++ {
				mh := makeHandler(ctrl)
				mh.EXPECT().Register(gomock.Any(), tl)
				handlers = append(handlers, mh)
			}

			s, err := New(tl, 4000, handlers...)
			if s == nil {
				t.Fatalf("Did not get `serve` back")
			}
			if err != nil {
				t.Errorf("Received unexpected error: %s", err.Error())
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Nanosecond)
			s.Serve(ctx)
			defer cancel()
		})
	}
}

type Fake0 struct {
	sample.UnimplementedSample0Server
}

func (f *Fake0) Register(s *grpc.Server, logger logr.Logger) error {
	sample.RegisterSample0Server(s, f)
	return nil
}

func (f *Fake0) Foo(ctx context.Context, req *sample.FooRequest) (*sample.FooResponse, error) {
	return &sample.FooResponse{Id: req.Id}, nil
}

type Fake1 struct {
	sample.UnimplementedSample1Server
}

func (f *Fake1) Register(s *grpc.Server, logger logr.Logger) error {
	sample.RegisterSample1Server(s, f)
	return nil
}

func (f *Fake1) Bar(ctx context.Context, req *sample.BarRequest) (*sample.BarResponse, error) {
	return &sample.BarResponse{Id: req.Id}, nil
}

func Test_GRPC_Server_OneClient(t *testing.T) {
	tl := testlogger.TestLogger{T: t}

	s, err := New(tl, 4000, &Fake0{})
	if err != nil {
		t.Fatalf("unexpected error from New: %s", err.Error())
	}
	if s == nil {
		t.Fatalf("unexpected nil from New")
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		s.Serve(ctx)
	}()

	cc, err := grpc.Dial(":4000", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		t.Fatalf("failed to dial client: %s", err.Error())
	}
	client := sample.NewSample0Client(cc)
	x := uuid.New().String()
	resp, err := client.Foo(context.Background(), &sample.FooRequest{Id: &sample.UUID{Value: x}})
	if err != nil {
		t.Fatalf("unexpecter error from Foo: %s", err.Error())
	}
	if resp.Id == nil {
		t.Errorf("returned Id is nil")
	}
	if resp.Id.Value != x {
		t.Errorf("returned incorrect Id value: expected '%s', actual: '%s'", x, resp.Id.Value)
	}
}

func Test_GRPC_Server_TwoClients(t *testing.T) {
	s, _ := New(testlogger.TestLogger{T: t}, 4000, &Fake0{}, &Fake1{})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		s.Serve(ctx)
	}()

	cc, _ := grpc.Dial(":4000", grpc.WithInsecure(), grpc.WithBlock())

	x0 := uuid.New().String()
	resp0, _ := sample.NewSample0Client(cc).Foo(context.Background(), &sample.FooRequest{Id: &sample.UUID{Value: x0}})
	if resp0.Id == nil {
		t.Errorf("returned Id is nil")
	} else if resp0.Id.Value != x0 {
		t.Errorf("returned incorrect Id value: expected '%s', actual: '%s'", x0, resp0.Id.Value)
	}

	x1 := uuid.New().String()
	resp1, _ := sample.NewSample1Client(cc).Bar(context.Background(), &sample.BarRequest{Id: &sample.UUID{Value: x1}})
	if resp1.Id == nil {
		t.Errorf("returned Id is nil")
	} else if resp1.Id.Value != x1 {
		t.Errorf("returned incorrect Id value: expected '%s', actual: '%s'", x1, resp1.Id.Value)
	}
}
