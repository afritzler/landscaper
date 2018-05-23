package util

import (
	"fmt"

	"github.com/mandelsoft/cmdint/pkg/cmdint"

	"github.com/afritzler/garden-examiner/cmd/gex/context"
	_ "github.com/afritzler/garden-examiner/pkg"
	"github.com/afritzler/garden-examiner/pkg/data"
)

type Handler interface {
	GetDefault(opts *cmdint.Options) *string
	RequireScan(string) bool
	MatchName(interface{}, string) (bool, error)
	Get(*context.Context, string) (interface{}, error)
	Iterator(ctx *context.Context, opts *cmdint.Options) (data.Iterator, error)
	Match(*context.Context, interface{}, *cmdint.Options) (bool, error)
	Add(*context.Context, interface{}) error
	Close(*context.Context) error
	Out(*context.Context)
}

/////////////////////////////////////////////////////////////////////////////
// Basic handler

type HandlerAdapter interface {
	GetAll(ctx *context.Context, opts *cmdint.Options) ([]interface{}, error)
	GetFilter() Filter
}

type BasicHandler struct {
	output Output
	elems  data.IndexedAccess
	impl   HandlerAdapter
}

func NewBasicHandler(o Output, impl HandlerAdapter) *BasicHandler {
	return (&BasicHandler{}).new(o, impl)
}

func (this *BasicHandler) new(o Output, impl HandlerAdapter) *BasicHandler {
	this.output = o
	this.elems = nil
	this.impl = impl
	return this
}

func NewBasicModeHandler(opts *cmdint.Options, outs Outputs, impl HandlerAdapter) (*BasicHandler, error) {
	o, err := outs.Create(opts)
	if err != nil {
		return nil, err
	}
	return (&BasicHandler{}).new(o, impl), nil
}

func (this *BasicHandler) RequireScan(name string) bool {
	return false
}

func (this *BasicHandler) GetDefault(opts *cmdint.Options) *string {
	return nil
}

func (this *BasicHandler) Iterator(ctx *context.Context, opts *cmdint.Options) (data.Iterator, error) {
	if this.elems == nil {
		elems, err := this.impl.GetAll(ctx, opts)
		if err != nil {
			return nil, err
		}
		this.elems = data.IndexedSliceAccess(elems)
	}
	return data.NewIndexedIterator(this.elems), nil
}

func (this *BasicHandler) Match(ctx *context.Context, e interface{}, opts *cmdint.Options) (bool, error) {
	return this.impl.GetFilter().Match(ctx, e, opts)
}

func (this *BasicHandler) Add(ctx *context.Context, e interface{}) error {
	return this.output.Add(ctx, e)
}

func (this *BasicHandler) Close(ctx *context.Context) error {
	return this.output.Close(ctx)
}
func (this *BasicHandler) Out(ctx *context.Context) {
	this.output.Out(ctx)
}

/////////////////////////////////////////////////////////////////////////////
// Self Handler

type SelfHandler interface {
	Handler
	HandlerAdapter
}

type BasicSelfHandler struct {
	self Handler
	BasicHandler
}

func NewBasicSelfHandler(o Output, self SelfHandler) *BasicSelfHandler {
	return (&BasicSelfHandler{}).new(o, self)
}

func (this *BasicSelfHandler) new(o Output, self SelfHandler) *BasicSelfHandler {
	this.BasicHandler.new(o, self)
	this.self = self
	return this
}

func NewBasicModeSelfHandler(opts *cmdint.Options, outs Outputs, self SelfHandler) (*BasicSelfHandler, error) {
	o, err := outs.Create(opts)
	if err != nil {
		return nil, err
	}
	return (&BasicSelfHandler{}).new(o, self), nil
}

func (this *BasicSelfHandler) Doit(opts *cmdint.Options) error {
	return Doit(opts, this.self)
}

/////////////////////////////////////////////////////////////////////////////
// Standard Command Logic

func DoitRaw(name_option string, opts *cmdint.Options, h Handler) error {
	ctx := context.Get(opts)
	name := ""
	if v := opts.GetOptionValue(name_option); v != nil {
		name = *v
	}
	if name == "" {
		if def := h.GetDefault(opts); def != nil {
			name = *def
		}
	}
	if name == "" {
		return fmt.Errorf("no element selected")
	}
	opts.Arguments = []string{name}
	return doDedicated(ctx, opts, h)
}

func Doit(opts *cmdint.Options, h Handler) error {
	ctx := context.Get(opts)

	if len(opts.Arguments) == 0 {
		if def := h.GetDefault(opts); def != nil {
			//fmt.Printf("DEFAULT: %s\n", *def)
			opts.Arguments = []string{*def}
		}
	}
	if len(opts.Arguments) > 0 && (len(opts.Arguments) != 1 || opts.Arguments[0] != "all") {
		return doDedicated(ctx, opts, h)
	} else {
		return doAll(ctx, opts, h)
	}
}

func doAll(ctx *context.Context, opts *cmdint.Options, h Handler) error {
	i, err := h.Iterator(ctx, opts)
	if err != nil {
		return err
	}
	for i.HasNext() {
		e := i.Next()
		ok, err := h.Match(ctx, e, opts)
		if err != nil {
			return err
		}
		if ok {
			err := h.Add(ctx, e)
			if err != nil {
				return err
			}
		}
	}
	h.Close(ctx)
	h.Out(ctx)
	return nil
}

func doDedicated(ctx *context.Context, opts *cmdint.Options, h Handler) error {
	for _, n := range opts.Arguments {
		if h.RequireScan(n) {
			i, err := h.Iterator(ctx, opts)
			if err != nil {
				return err
			}
			for _, n := range opts.Arguments {
				if !h.RequireScan(n) {
					e, err := h.Get(ctx, n)
					if err != nil {
						return err
					}
					if e == nil {
						return fmt.Errorf("'%s' not found", n)
					}
					ok, err := h.Match(ctx, e, opts)
					if err != nil {
						return err
					}
					if ok {
						err := h.Add(ctx, e)
						if err != nil {
							return err
						}
					}
				} else {
					//fmt.Printf("LOOKUP %s\n", n)
					found := false
					i, err = h.Iterator(ctx, opts)
					if err != nil {
						return err
					}
					for i.HasNext() {
						e := i.Next()
						ok, err := h.Match(ctx, e, opts)
						if err != nil {
							return err
						}
						//fmt.Printf("  check %s: %s\n", e.(gube.Shoot).GetName(), ok)
						if ok {
							ok, err := h.MatchName(e, n)
							if err != nil {
								return err
							}
							if ok {
								err := h.Add(ctx, e)
								if err != nil {
									return err
								}
								found = true
							}
						}
					}
					if !found {
						return fmt.Errorf("'%s' not found", n)
					}
				}
			}
			h.Out(ctx)
			return nil
		}
	}

	for _, n := range opts.Arguments {
		e, err := h.Get(ctx, n)
		if err != nil {
			return err
		}
		err = h.Add(ctx, e)
		if err != nil {
			return err
		}
	}
	h.Close(ctx)
	h.Out(ctx)
	return nil
}
