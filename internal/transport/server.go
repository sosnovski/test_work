package transport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"sync"

	"github.com/go-playground/validator/v10"

	"mascot/internal/handlers"
)

const version = "2.0"

var (
	ctxElement = reflect.TypeOf((*context.Context)(nil)).Elem()
	errElement = reflect.TypeOf((*error)(nil)).Elem()
)

type ServerOption func(server *Server)

type Server struct {
	handlers    sync.Map
	middlewares []MiddlewareFunc
	validate    *validator.Validate
}

func NewServer(options ...ServerOption) *Server {
	s := &Server{}
	for _, option := range options {
		option(s)
	}
	return s
}

func (s *Server) RegisterServices(namesAndFuncs ...interface{}) error {
	if len(namesAndFuncs)%2 != 0 {
		return errors.New("odd number of arguments")
	}
	for i := 0; i <= len(namesAndFuncs)-2; {
		funcIndex := i + 1
		name := namesAndFuncs[i]
		fun := namesAndFuncs[funcIndex]

		kindName := reflect.TypeOf(name).Kind()
		kindFun := reflect.TypeOf(fun).Kind()

		if kindName != reflect.String {
			return fmt.Errorf("%d argument must be string", i)
		}

		if kindFun != reflect.Func {
			return fmt.Errorf("%d argument must be func", funcIndex)
		}

		if err := s.RegisterService(name.(string), fun); err != nil {
			return err
		}

		i += 2
	}

	return nil
}

func (s *Server) RegisterService(name string, fun interface{}) error {
	valFun := reflect.ValueOf(fun)
	valType := reflect.TypeOf(fun)

	if valType.Kind() != reflect.Func {
		return fmt.Errorf("handler must be a reflect.Func type")
	}

	h := &handler{
		fun:         valFun,
		argCount:    1,
		resultCount: 1,
	}

	switch valType.NumIn() {
	case 1:
	case 2:
		h.argType = valType.In(1)
		h.argIsValue = valType.Kind() != reflect.Ptr
		h.argCount = 2
	default:
		return fmt.Errorf("number of return values must be 1 or 2")
	}

	firstArg := valType.In(0)
	if firstArg.Kind() != reflect.Interface || !firstArg.Implements(ctxElement) {
		return fmt.Errorf("first argument in handler must be implement context.Context")
	}

	switch valType.NumOut() {
	case 1:
	case 2:
		h.resultType = valType.Out(0)
		h.resultCount = 2
		h.errIndex = 1
	default:
		return fmt.Errorf("number of return values must be 1 or 2")
	}

	lastRes := valType.Out(valType.NumOut() - 1)
	if lastRes.Kind() != reflect.Interface || !lastRes.Implements(errElement) {
		return fmt.Errorf("last return value must implement error interface")
	}

	if _, ok := s.handlers.Load(name); ok {
		return fmt.Errorf("handler with name %s already registered", name)
	}

	s.handlers.Store(name, h)
	return nil
}

func (s *Server) serveFunc() func(httpReq *http.Request) (resp *ServerResponse) {
	n := len(s.middlewares)
	handler := s.handle
	for i := n - 1; i >= 0; i-- {
		handler = s.middlewares[i](handler)
	}

	return func(httpReq *http.Request) (resp *ServerResponse) {
		resp = &ServerResponse{Jsonrpc: version}
		req := &ServerRequest{}
		if err := json.NewDecoder(httpReq.Body).Decode(req); err != nil {
			resp.Error = handlers.NewError(handlers.ErrParse, "parse error")
			return
		}

		if httpReq.Method != http.MethodPost {
			resp.Error = handlers.NewError(handlers.ErrMethodNotFound, "http method not found")
			return
		}

		resp.Id = req.Id
		handler(httpReq.Context(), req, resp)
		return
	}
}

func (s *Server) handle(ctx context.Context, req *ServerRequest, resp *ServerResponse) {
	if req.Method == "" {
		resp.Error = handlers.NewError(handlers.ErrInvalidRequest, "method is missing")
		return
	}

	val, ok := s.handlers.Load(req.Method)
	if !ok {
		resp.Error = handlers.NewError(handlers.ErrMethodNotFound, "method not found")
		return
	}

	h, ok := val.(*handler)
	if !ok {
		resp.Error = handlers.NewError(handlers.ErrInternalError, "internal error")
		return
	}

	values := make([]reflect.Value, h.argCount)

	values[0] = reflect.ValueOf(ctx)
	if h.argType != nil {
		if req.Params == nil {
			resp.Error = handlers.NewError(handlers.ErrInvalidRequest, "params is missing")
			return
		}

		var request reflect.Value
		if h.argIsValue {
			request = reflect.New(h.argType)
		} else {
			request = reflect.New(h.argType.Elem())
		}

		if err := json.Unmarshal(req.Params, request.Interface()); err != nil {
			resp.Error = handlers.NewError(handlers.ErrInvalidRequest, "invalid request")
			return
		}

		if h.argIsValue {
			request = request.Elem()
		}

		if s.validate != nil {
			if err := s.validate.Struct(request.Interface()); err != nil {
				resp.Error = handlers.NewError(handlers.ErrInvalidParams, err.Error())
				return
			}
		}

		values[1] = request
	}

	results := h.fun.Call(values)

	if h.resultType != nil && !results[0].IsNil() {
		resp.Result = results[0].Interface()
	}

	refErr := results[h.errIndex]
	if refErr.IsNil() {
		if h.resultType == nil {
			resp.Result = struct{}{}
		}
		return
	}

	var respErr *handlers.Error
	err := refErr.Interface().(error)

	if errors.As(err, &respErr) {
		resp.Error = respErr
	} else {
		resp.Error = &handlers.Error{
			Code:    handlers.ErrDefaultServerError,
			Message: err.Error(),
		}
	}

	return
}

func (s *Server) HandleFunc() http.HandlerFunc {
	serveFunc := s.serveFunc()

	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := serveFunc(req)
		b, err := json.Marshal(resp)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err.Error())
		}

		_, err = w.Write(b)
	}
}

func (s *Server) UseMiddlewares(middlewares ...MiddlewareFunc) *Server {
	s.middlewares = middlewares
	return s
}

func WithUseValidator() ServerOption {
	return func(server *Server) {
		server.validate = validator.New()
	}
}

type handler struct {
	fun         reflect.Value
	argType     reflect.Type
	argIsValue  bool
	argCount    int
	resultType  reflect.Type
	resultCount int
	errIndex    int
}

type ServerRequest struct {
	Id     int             `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

type ServerResponse struct {
	Jsonrpc string          `json:"jsonrpc"`
	Id      int             `json:"id"`
	Result  interface{}     `json:"result,omitempty"`
	Error   *handlers.Error `json:"error,omitempty"`
}
