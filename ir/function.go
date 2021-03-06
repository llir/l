package ir

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/llir/l/internal/enc"
	"github.com/llir/l/ir/enum"
	"github.com/llir/l/ir/types"
	"github.com/llir/l/ir/value"
	"github.com/pkg/errors"
)

// === [ Functions ] ===========================================================

// Function is an LLVM IR function.
type Function struct {
	// Function signature.
	Sig *types.FuncType
	// Function name (without '@' prefix).
	GlobalName string
	// Function parameters.
	Params []*Param
	// Basic blocks.
	Blocks []*BasicBlock

	// extra.

	// Pointer type to function, including an optional address space. If Typ is
	// nil, the first invocation of Type stores a pointer type with Sig as
	// element.
	Typ *types.PointerType
	// (optional) Linkage.
	Linkage enum.Linkage
	// (optional) Preemption; zero value if not present.
	Preemption enum.Preemption
	// (optional) Visibility; zero value if not present.
	Visibility enum.Visibility
	// (optional) DLL storage class; zero value if not present.
	DLLStorageClass enum.DLLStorageClass
	// (optional) Calling convention; zero value if not present.
	CallingConv enum.CallingConv
	// (optional) Return attributes.
	ReturnAttrs []enum.ReturnAttribute
	// (optional) Unnamed address.
	UnnamedAddr enum.UnnamedAddr
	// (optional) Function attributes.
	FuncAttrs []enum.FuncAttribute
	// (optional) Section; nil if not present.
	Section string
	// (optional) Comdat; nil if not present.
	Comdat *ComdatDef
	// (optional) Garbage collection; empty if not present.
	GC string
	// (optional) Prefix; nil if not present.
	Prefix Constant
	// (optional) Prologue; nil if not present.
	Prologue Constant
	// (optional) Personality; nil if not present.
	Personality Constant
	// (optional) Use list orders.
	// TODO: add support for UseListOrder.
	//UseListOrders []*UseListOrder
	// (optional) Metadata attachments.
	// TODO: add support for metadata.
	//Metadata []*metadata.MetadataAttachment
}

// TODO: decide whether to have the function name parameter be the first
// argument (to be consistent with NewGlobal) or after retType (to be consistent
// with order of occurence in LLVM IR syntax).

// NewFunction returns a new function based on the given function name, return
// type and function parameters.
func NewFunction(name string, retType types.Type, params ...*Param) *Function {
	panic("not yet implemented")
	/*
		paramTypes := make([]types.Type, len(params))
		for i, param := range f.Params {
			paramType[i] = param.Type()
		}
		sig := types.NewFunc(f.RetType, paramTypes...)
		return &Function{Sig: sig, GlobalName: name, Params: params}
	*/
}

// String returns the LLVM syntax representation of the function as a type-value
// pair.
func (f *Function) String() string {
	return fmt.Sprintf("%v %v", f.Type(), f.Ident())
}

// Type returns the type of the function.
func (f *Function) Type() types.Type {
	// Cache type if not present.
	if f.Typ == nil {
		f.Typ = types.NewPointer(f.Sig)
	}
	return f.Typ
}

// Ident returns the identifier associated with the function.
func (f *Function) Ident() string {
	return enc.Global(f.GlobalName)
}

// Name returns the name of the function.
func (f *Function) Name() string {
	return f.GlobalName
}

// SetName sets the name of the function.
func (f *Function) SetName(name string) {
	f.GlobalName = name
}

// Def returns the LLVM syntax representation of the function definition or
// declaration.
func (f *Function) Def() string {
	// "declare" MetadataAttachments OptExternLinkage FunctionHeader
	// "define" OptLinkage FunctionHeader MetadataAttachments FunctionBody
	buf := &strings.Builder{}
	if len(f.Blocks) == 0 {
		// Function declaration.
		//
		//    "declare" MetadataAttachments OptExternLinkage FunctionHeader
		buf.WriteString("declare")
		// TODO: add metadata support.
		//for _, md := range f.Metadata {
		//	fmt.Fprintf(buf, " %v", md)
		//}
		if f.Linkage != enum.LinkageNone {
			fmt.Fprintf(buf, " %v", f.Linkage)
		}
		buf.WriteString(headerString(f))
		return buf.String()
	}
	// Function definition.
	//
	//    "define" OptLinkage FunctionHeader MetadataAttachments FunctionBody
	buf.WriteString("define")
	if f.Linkage != enum.LinkageNone {
		fmt.Fprintf(buf, " %v", f.Linkage)
	}
	buf.WriteString(headerString(f))
	// TODO: add metadata support.
	//for _, md := range f.Metadata {
	//	fmt.Fprintf(buf, " %v", md)
	//}
	fmt.Fprintf(buf, " %v", bodyString(f))
	return buf.String()
}

// AssignIDs assigns IDs to unnamed local variables.
func (f *Function) AssignIDs() error {
	if len(f.Blocks) == 0 {
		return nil
	}
	id := 0
	names := make(map[string]value.Value)
	setName := func(n value.Named) error {
		got := n.Name()
		if isUnnamed(got) {
			name := strconv.Itoa(id)
			n.SetName(name)
			names[name] = n
			id++
		} else if isLocalID(got) {
			want := strconv.Itoa(id)
			if want != got {
				return errors.Errorf("invalid local ID in function %q, expected %s, got %s", enc.Global(f.GlobalName), enc.Local(want), enc.Local(got))
			}
			id++
		} else {
			// already named; nothing to do.
		}
		return nil
	}
	for _, param := range f.Params {
		// Assign local IDs to unnamed parameters of function definitions.
		if err := setName(param); err != nil {
			return errors.WithStack(err)
		}
	}
	for _, block := range f.Blocks {
		// Assign local IDs to unnamed basic blocks.
		if err := setName(block); err != nil {
			return errors.WithStack(err)
		}
		for _, inst := range block.Insts {
			n, ok := inst.(value.Named)
			if !ok {
				continue
			}
			// Skip void instructions.
			// TODO: Check if any other value instructions than call may have void
			// type.
			if isVoidValue(n) {
				continue
			}
			// Assign local IDs to unnamed local variables.
			if err := setName(n); err != nil {
				return errors.WithStack(err)
			}
		}
		n, ok := block.Term.(value.Named)
		if !ok {
			continue
		}
		if isVoidValue(n) {
			continue
		}
		if err := setName(n); err != nil {
			return errors.WithStack(err)
		}
	}
	return nil
}

// ### [ Helper functions ] ####################################################

// headerString returns the string representation of the function header.
func headerString(hdr *Function) string {
	// OptPreemptionSpecifier OptVisibility OptDLLStorageClass OptCallingConv
	// ReturnAttrs Type GlobalIdent "(" Params ")" OptUnnamedAddr FuncAttrs
	// OptSection OptComdat OptGC OptPrefix OptPrologue OptPersonality
	buf := &strings.Builder{}
	if hdr.Preemption != enum.PreemptionNone {
		fmt.Fprintf(buf, " %v", hdr.Preemption)
	}
	if hdr.Visibility != enum.VisibilityNone {
		fmt.Fprintf(buf, " %v", hdr.Visibility)
	}
	if hdr.DLLStorageClass != enum.DLLStorageClassNone {
		fmt.Fprintf(buf, " %v", hdr.DLLStorageClass)
	}
	if hdr.CallingConv != enum.CallingConvNone {
		fmt.Fprintf(buf, " %v", hdr.CallingConv)
	}
	for _, attr := range hdr.ReturnAttrs {
		fmt.Fprintf(buf, " %v", attr)
	}
	fmt.Fprintf(buf, " %v", hdr.Sig.RetType)
	fmt.Fprintf(buf, " %v(", enc.Global(hdr.GlobalName))
	for i, param := range hdr.Params {
		if i != 0 {
			buf.WriteString(", ")
		}
		buf.WriteString(param.Def())
	}
	if hdr.Sig.Variadic {
		if len(hdr.Params) > 0 {
			buf.WriteString(", ")
		}
		buf.WriteString("...")
	}
	buf.WriteString(")")
	if hdr.UnnamedAddr != enum.UnnamedAddrNone {
		fmt.Fprintf(buf, " %v", hdr.UnnamedAddr)
	}
	for _, attr := range hdr.FuncAttrs {
		fmt.Fprintf(buf, " %v", attr)
	}
	if len(hdr.Section) > 0 {
		fmt.Fprintf(buf, " section %v", enc.Quote([]byte(hdr.Section)))
	}
	if hdr.Comdat != nil {
		fmt.Fprintf(buf, " %v", hdr.Comdat)
	}
	if len(hdr.GC) > 0 {
		fmt.Fprintf(buf, " gc %v", enc.Quote([]byte(hdr.GC)))
	}
	if hdr.Prefix != nil {
		fmt.Fprintf(buf, " prefix %v", hdr.Prefix)
	}
	if hdr.Prologue != nil {
		fmt.Fprintf(buf, " prologue %v", hdr.Prologue)
	}
	if hdr.Personality != nil {
		fmt.Fprintf(buf, " personality %v", hdr.Personality)
	}
	return buf.String()
}

// bodyString returns the string representation of the function body.
func bodyString(body *Function) string {
	// "{" BasicBlockList UseListOrders "}"
	buf := &strings.Builder{}
	buf.WriteString("{\n")
	for _, block := range body.Blocks {
		fmt.Fprintf(buf, "%v\n", block.Def())
	}
	// TODO: add support for use list orders.
	//for _, useList := range body.UseListOrders {
	//	fmt.Fprintf(buf, "%v\n", useList)
	//}
	buf.WriteString("}")
	return buf.String()
}

// isVoidValue reports whether the given named value is a non-value (i.e. a call
// instruction or invoke terminator with void-return type).
func isVoidValue(n value.Named) bool {
	switch n.(type) {
	case *InstCall, *TermInvoke:
		return n.Type().Equal(types.Void)
	}
	return false
}
