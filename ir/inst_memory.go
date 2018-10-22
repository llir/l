package ir

import (
	"fmt"

	"github.com/llir/l/internal/enc"
	"github.com/llir/l/ir/enum"
	"github.com/llir/l/ir/types"
	"github.com/llir/l/ir/value"
)

// --- [ Memory instructions ] -------------------------------------------------

// ~~~ [ alloca ] ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// InstAlloca is an LLVM IR alloca instruction.
type InstAlloca struct {
	// Name of local variable associated with the result.
	LocalName string
	// Element type.
	ElemType types.Type
	// (optional) Number of elements; nil if not present.
	NElems value.Value

	// extra.

	// Type of result produced by the instruction, including an optional address
	// space.
	Typ *types.PointerType
	// (optional) In-alloca.
	InAlloca bool
	// (optional) Swift error.
	SwiftError bool
	// (optional) Alignment; zero if not present.
	Alignment int
	// (optional) Metadata.
	// TODO: add metadata.
}

// NewAlloca returns a new alloca instruction based on the given element type.
func NewAlloca(elemType types.Type) *InstAlloca {
	return &InstAlloca{ElemType: elemType}
}

// String returns the LLVM syntax representation of the instruction as a
// type-value pair.
func (inst *InstAlloca) String() string {
	return fmt.Sprintf("%v %v", inst.Type(), inst.Ident())
}

// Type returns the type of the instruction.
func (inst *InstAlloca) Type() types.Type {
	// Cache type if not present.
	if inst.Typ == nil {
		inst.Typ = types.NewPointer(inst.ElemType)
	}
	return inst.Typ
}

// Ident returns the identifier associated with the instruction.
func (inst *InstAlloca) Ident() string {
	return enc.Local(inst.LocalName)
}

// Name returns the name of the instruction.
func (inst *InstAlloca) Name() string {
	return inst.LocalName
}

// SetName sets the name of the instruction.
func (inst *InstAlloca) SetName(name string) {
	inst.LocalName = name
}

// ~~~ [ load ] ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// InstLoad is an LLVM IR load instruction.
type InstLoad struct {
	// Name of local variable associated with the result.
	LocalName string
	// Source address.
	Src value.Value

	// extra.

	// Type of result produced by the instruction.
	Typ types.Type
	// (optional) Atomic.
	Atomic bool
	// (optional) Volatile.
	Volatile bool
	// (optional) Sync scope; empty if not present.
	SyncScope string
	// (optional) Atomic memory ordering constraints; zero if not present.
	Ordering enum.AtomicOrdering
	// (optional) Alignment; zero if not present.
	Alignment int
	// (optional) Metadata.
	// TODO: add metadata.
}

// NewLoad returns a new load instruction based on the given source address.
func NewLoad(src value.Value) *InstLoad {
	return &InstLoad{Src: src}
}

// String returns the LLVM syntax representation of the instruction as a
// type-value pair.
func (inst *InstLoad) String() string {
	return fmt.Sprintf("%v %v", inst.Type(), inst.Ident())
}

// Type returns the type of the instruction.
func (inst *InstLoad) Type() types.Type {
	// Cache type if not present.
	if inst.Typ == nil {
		t, ok := inst.Src.Type().(*types.PointerType)
		if !ok {
			panic(fmt.Errorf("invalid source type; expected *types.PointerType, got %T", inst.Src.Type()))
		}
		inst.Typ = t.ElemType
	}
	return inst.Typ
}

// Ident returns the identifier associated with the instruction.
func (inst *InstLoad) Ident() string {
	return enc.Local(inst.LocalName)
}

// Name returns the name of the instruction.
func (inst *InstLoad) Name() string {
	return inst.LocalName
}

// SetName sets the name of the instruction.
func (inst *InstLoad) SetName(name string) {
	inst.LocalName = name
}

// ~~~ [ store ] ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// InstStore is an LLVM IR store instruction.
type InstStore struct {
	// Source value.
	Src value.Value
	// Destination address.
	Dst value.Value

	// extra.

	// (optional) Atomic.
	Atomic bool
	// (optional) Volatile.
	Volatile bool
	// (optional) Sync scope; empty if not present.
	SyncScope string
	// (optional) Atomic memory ordering constraints; zero if not present.
	Ordering enum.AtomicOrdering
	// (optional) Alignment; zero if not present.
	Alignment int
	// (optional) Metadata.
	// TODO: add metadata.
}

// NewStore returns a new store instruction based on the given source value and
// destination address.
func NewStore(src, dst value.Value) *InstStore {
	return &InstStore{Src: src, Dst: dst}
}

// ~~~ [ fence ] ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// InstFence is an LLVM IR fence instruction.
type InstFence struct {
	// Atomic memory ordering constraints.
	Ordering enum.AtomicOrdering

	// extra.

	// (optional) Sync scope; empty if not present.
	SyncScope string
	// (optional) Metadata.
	// TODO: add metadata.
}

// NewFence returns a new fence instruction based on the given atomic ordering.
func NewFence(ordering enum.AtomicOrdering) *InstFence {
	return &InstFence{Ordering: ordering}
}

// ~~~ [ cmpxchg ] ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// InstCmpXchg is an LLVM IR cmpxchg instruction.
type InstCmpXchg struct {
	// Name of local variable associated with the result.
	LocalName string
	// Address to read from, compare against and store to.
	Ptr value.Value
	// Value to compare against.
	Cmp value.Value
	// New value to store.
	New value.Value
	// Atomic memory ordering constraints on success.
	Success enum.AtomicOrdering
	// Atomic memory ordering constraints on failure.
	Failure enum.AtomicOrdering

	// extra.

	// Type of result produced by the instruction; the first field of the struct
	// holds the old value, and the second field indicates success.
	Typ *types.StructType
	// (optional) Weak.
	Weak bool
	// (optional) Volatile.
	Volatile bool
	// (optional) Sync scope; empty if not present.
	SyncScope string
	// (optional) Metadata.
	// TODO: add metadata.
}

// NewCmpXchg returns a new cmpxchg instruction based on the given address,
// value to compare against, new value to store, and atomic orderings for
// success and failure.
func NewCmpXchg(ptr, cmp, new value.Value, success, failure enum.AtomicOrdering) *InstCmpXchg {
	return &InstCmpXchg{Ptr: ptr, Cmp: cmp, New: new, Success: success, Failure: failure}
}

// String returns the LLVM syntax representation of the instruction as a
// type-value pair.
func (inst *InstCmpXchg) String() string {
	return fmt.Sprintf("%v %v", inst.Type(), inst.Ident())
}

// Type returns the type of the instruction.
func (inst *InstCmpXchg) Type() types.Type {
	// Cache type if not present.
	if inst.Typ == nil {
		oldType := inst.New.Type()
		inst.Typ = types.NewStruct(oldType, types.I1)
	}
	return inst.Typ
}

// Ident returns the identifier associated with the instruction.
func (inst *InstCmpXchg) Ident() string {
	return enc.Local(inst.LocalName)
}

// Name returns the name of the instruction.
func (inst *InstCmpXchg) Name() string {
	return inst.LocalName
}

// SetName sets the name of the instruction.
func (inst *InstCmpXchg) SetName(name string) {
	inst.LocalName = name
}

// ~~~ [ atomicrmw ] ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// InstAtomicRMW is an LLVM IR atomicrmw instruction.
type InstAtomicRMW struct {
	// Name of local variable associated with the result.
	LocalName string
	// Atomic operation.
	Op enum.AtomicOp
	// Destination address.
	Dst value.Value
	// Operand.
	X value.Value
	// Atomic memory ordering constraints.
	Ordering enum.AtomicOrdering

	// extra.

	// Type of result produced by the instruction.
	Typ types.Type
	// (optional) Volatile.
	Volatile bool
	// (optional) Sync scope; empty if not present.
	SyncScope string
	// (optional) Metadata.
	// TODO: add metadata.
}

// NewAtomicRMW returns a new atomicrmw instruction based on the given atomic
// operation, destination address, operand and atomic ordering.
func NewAtomicRMW(op enum.AtomicOp, dst, x value.Value, ordering enum.AtomicOrdering) *InstAtomicRMW {
	return &InstAtomicRMW{Op: op, Dst: dst, X: x, Ordering: ordering}
}

// String returns the LLVM syntax representation of the instruction as a
// type-value pair.
func (inst *InstAtomicRMW) String() string {
	return fmt.Sprintf("%v %v", inst.Type(), inst.Ident())
}

// Type returns the type of the instruction.
func (inst *InstAtomicRMW) Type() types.Type {
	// Cache type if not present.
	if inst.Typ == nil {
		t, ok := inst.Dst.Type().(*types.PointerType)
		if !ok {
			panic(fmt.Errorf("invalid destination type; expected *types.PointerType, got %T", inst.Dst.Type()))
		}
		inst.Typ = t.ElemType
	}
	return inst.Typ
}

// Ident returns the identifier associated with the instruction.
func (inst *InstAtomicRMW) Ident() string {
	return enc.Local(inst.LocalName)
}

// Name returns the name of the instruction.
func (inst *InstAtomicRMW) Name() string {
	return inst.LocalName
}

// SetName sets the name of the instruction.
func (inst *InstAtomicRMW) SetName(name string) {
	inst.LocalName = name
}

// ~~~ [ getelementptr ] ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

// InstGetElementPtr is an LLVM IR getelementptr instruction.
type InstGetElementPtr struct {
	// Name of local variable associated with the result.
	LocalName string
	// Element type.
	ElemType types.Type
	// Source address.
	Src value.Value
	// Element indicies.
	Indices []value.Value

	// extra.

	// Type of result produced by the instruction.
	Typ types.Type
	// (optional) In-bounds.
	InBounds bool
	// (optional) Metadata.
	// TODO: add metadata.
}

// NewGetElementPtr returns a new getelementptr instruction based on the given
// element type, source address and element indices.
func NewGetElementPtr(elemType types.Type, src value.Value, indices ...value.Value) *InstGetElementPtr {
	return &InstGetElementPtr{ElemType: elemType, Src: src, Indices: indices}
}

// String returns the LLVM syntax representation of the instruction as a
// type-value pair.
func (inst *InstGetElementPtr) String() string {
	return fmt.Sprintf("%v %v", inst.Type(), inst.Ident())
}

// Type returns the type of the instruction.
func (inst *InstGetElementPtr) Type() types.Type {
	// Cache type if not present.
	if inst.Typ == nil {
		inst.Typ = types.NewPointer(inst.ElemType)
	}
	return inst.Typ
}

// Ident returns the identifier associated with the instruction.
func (inst *InstGetElementPtr) Ident() string {
	return enc.Local(inst.LocalName)
}

// Name returns the name of the instruction.
func (inst *InstGetElementPtr) Name() string {
	return inst.LocalName
}

// SetName sets the name of the instruction.
func (inst *InstGetElementPtr) SetName(name string) {
	inst.LocalName = name
}
