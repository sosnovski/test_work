package transport

import (
	"context"
	"testing"
)

func TestServer_RegisterServices(t *testing.T) {
	t.Parallel()
	handleFunc := func(ctx context.Context) error { return nil }

	type args struct {
		namesAndFuncs []interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "even number of arguments with with correct types",
			args: args{
				namesAndFuncs: []interface{}{"name", handleFunc, "anotherName", handleFunc},
			},
			wantErr: false,
		},
		{
			name: "name with wrong type",
			args: args{
				namesAndFuncs: []interface{}{1, handleFunc},
			},
			wantErr: true,
		},
		{
			name: "func with wrong type",
			args: args{
				namesAndFuncs: []interface{}{"name", 1},
			},
			wantErr: true,
		},
		{
			name: "odd number of arguments",
			args: args{
				namesAndFuncs: []interface{}{"name", handleFunc, "anotherName"},
			},
			wantErr: true,
		},
		{
			name: "RegisterService return error",
			args: args{
				namesAndFuncs: []interface{}{"name", func() error { return nil }},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			s := NewServer()
			if err := s.RegisterServices(tt.args.namesAndFuncs...); (err != nil) != tt.wantErr {
				t.Errorf("RegisterServices() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_server_RegisterService(t *testing.T) {
	t.Parallel()
	type args struct {
		name string
		fun  interface{}
	}
	tests := []struct {
		name    string
		args    args
		server  *Server
		wantErr bool
	}{
		{
			name: "without argument and result",
			args: args{
				name: "name",
				fun:  func(ctx context.Context) error { return nil },
			},
			server:  NewServer(),
			wantErr: false,
		},
		{
			name: "without argument with result",
			args: args{
				name: "name",
				fun:  func(ctx context.Context) (interface{}, error) { return nil, nil },
			},
			server:  NewServer(),
			wantErr: false,
		},
		{
			name: "three result",
			args: args{
				name: "name",
				fun:  func(ctx context.Context) (interface{}, interface{}, error) { return nil, nil, nil },
			},
			server:  NewServer(),
			wantErr: true,
		},
		{
			name: "without result with argument",
			args: args{
				name: "name",
				fun:  func(ctx context.Context, val interface{}) error { return nil },
			},
			server:  NewServer(),
			wantErr: false,
		},
		{
			name: "with argument and result",
			args: args{
				name: "name",
				fun:  func(ctx context.Context, val interface{}) (interface{}, error) { return nil, nil },
			},
			server:  NewServer(),
			wantErr: false,
		},
		{
			name: "two argument",
			args: args{
				name: "name",
				fun:  func(ctx context.Context, val interface{}, val2 interface{}) error { return nil },
			},
			server:  NewServer(),
			wantErr: true,
		},
		{
			name: "argument without context",
			args: args{
				name: "name",
				fun:  func(val interface{}, val2 interface{}) error { return nil },
			},
			server:  NewServer(),
			wantErr: true,
		},
		{
			name: "without all arguments",
			args: args{
				name: "name",
				fun:  func() error { return nil },
			},
			server:  NewServer(),
			wantErr: true,
		},
		{
			name: "without error result",
			args: args{
				name: "name",
				fun:  func(ctx context.Context) interface{} { return nil },
			},
			server:  NewServer(),
			wantErr: true,
		},
		{
			name: "err is not second argument",
			args: args{
				name: "name",
				fun:  func(ctx context.Context) (error, interface{}) { return nil, nil },
			},
			server:  NewServer(),
			wantErr: true,
		},
		{
			name: "err is not second result with one result",
			args: args{
				name: "name",
				fun:  func(ctx context.Context) error { return nil },
			},
			server:  NewServer(),
			wantErr: false,
		},
		{
			name: "fun is not function type",
			args: args{
				name: "name",
				fun:  1,
			},
			server:  NewServer(),
			wantErr: true,
		},
		{
			name: "handler already registered",
			args: args{
				name: "name",

				fun: func(ctx context.Context) error { return nil },
			},
			server: func() *Server {
				s := NewServer()
				_ = s.RegisterServices("name", func(ctx context.Context) error { return nil })
				return s
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.server.RegisterService(tt.args.name, tt.args.fun); (err != nil) != tt.wantErr {
				t.Errorf("RegisterService() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
