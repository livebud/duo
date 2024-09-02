package scope

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/tdewolff/parse/v2/js"
)

func New() *Scope {
	return &Scope{}
}

type Scope struct {
	parent  *Scope
	symbols []*Symbol

	// These are set while traversing the AST
	IsExported    bool `json:"is_exported,omitempty"`
	IsDeclaration bool `json:"is_declaration,omitempty"`
	IsMutable     bool `json:"is_mutable,omitempty"`
}

func (s *Scope) Use(name string) *Symbol {
	if sym, ok := s.LookupByName(name); ok {
		// TODO: a bitmask could make this more terse
		if !sym.isDeclared {
			sym.isDeclared = s.IsDeclaration
		}
		if !sym.isExported {
			sym.isExported = s.IsExported
		}
		if !sym.isMutable {
			sym.isMutable = s.IsMutable
		}
		return sym
	}
	sym := &Symbol{
		Name:       name,
		isDeclared: s.IsDeclaration,
		isExported: s.IsExported,
		isMutable:  s.IsMutable,
	}
	s.symbols = append(s.symbols, sym)
	return sym
}

// Declare a new variable with a unique id
func (s *Scope) Declare(id, name string) (*Symbol, error) {
	if _, ok := s.LookupByID(id); ok {
		return nil, fmt.Errorf("symbol with id %q already declared", id)
	}
	name = s.FindFree(name)
	sym := &Symbol{
		Name:       name,
		ID:         id,
		isDeclared: s.IsDeclaration,
		isExported: s.IsExported,
		isMutable:  s.IsMutable,
	}
	s.symbols = append(s.symbols, sym)
	return sym, nil
}

func (s *Scope) LookupByName(name string) (*Symbol, bool) {
	for _, sym := range s.symbols {
		if sym.Name == name {
			return sym, true
		}
	}
	if s.parent != nil {
		return s.parent.LookupByName(name)
	}
	return nil, false
}

func (s *Scope) LookupByID(id string) (*Symbol, bool) {
	for _, sym := range s.symbols {
		if sym.ID == id {
			return sym, true
		}
	}
	if s.parent != nil {
		return s.parent.LookupByID(id)
	}
	return nil, false
}

func (s *Scope) New() *Scope {
	return &Scope{
		parent:        s,
		IsExported:    s.IsExported,
		IsDeclaration: s.IsDeclaration,
		IsMutable:     s.IsMutable,
	}
}

func (s *Scope) Parent() *Scope {
	return s.parent
}

func (s *Scope) Clone() *Scope {
	var parent *Scope
	if s.parent != nil {
		parent = s.parent.Clone()
	}
	clone := &Scope{
		parent:        parent,
		symbols:       s.symbols,
		IsExported:    s.IsExported,
		IsDeclaration: s.IsDeclaration,
		IsMutable:     s.IsMutable,
	}
	return clone
}

// FindFree returns a name that is not already declared in the scope.
func (s *Scope) FindFree(name string) string {
	if _, ok := s.LookupByName(name); !ok {
		return name
	}
	for i := 0; ; i++ {
		if _, ok := s.LookupByName(fmt.Sprintf("%s%d", name, i)); !ok {
			return fmt.Sprintf("%s%d", name, i)
		}
	}
}

// func (s *Scope) Declared() []*Symbol {
// 	var symbols []*Symbol
// 	for _, sym := range s.declared {
// 		symbols = append(symbols, sym)
// 	}
// 	sort.Slice(symbols, func(i, j int) bool {
// 		return symbols[i].Name < symbols[j].Name
// 	})
// 	return symbols
// }

// func (s *Scope) all() [][]*Symbol {
// 	var symbols [][]*Symbol
// 	if s.parent != nil {
// 		symbols = append(symbols, s.parent.all()...)
// 	}
// 	symbols = append(symbols, s.Declared())
// 	return symbols
// }

// func (s *Scope) MarshalJSON() ([]byte, error) {
// 	return json.Marshal(s.all())
// }

func (s *Scope) String() string {
	str := new(strings.Builder)
	if s.parent != nil {
		str.WriteString(s.parent.String())
		str.WriteString("\n")
	}
	for _, symbol := range s.symbols {
		str.WriteString(symbol.String())
		str.WriteString("\n")
	}
	return str.String()
}

type Symbol struct {
	Name       string  `json:"name,omitempty"`
	ID         string  `json:"id,omitempty"`
	Import     *Import // nil if not an import
	isDeclared bool
	isExported bool
	isMutable  bool
}

type Import struct {
	Path    string
	Default bool
}

func (s *Symbol) IsDeclared() bool {
	return s.isDeclared
}

func (s *Symbol) IsExported() bool {
	return s.isExported
}

func (s *Symbol) IsMutable() bool {
	return s.isMutable
}

func (s *Symbol) String() string {
	w := new(strings.Builder)
	w.WriteString(strconv.Quote(s.Name))
	if s.isDeclared {
		w.WriteString(" declared")
	}
	if s.isExported {
		w.WriteString(" exported")
	}
	if s.isMutable {
		w.WriteString(" mutable")
	}
	if s.Import != nil {
		w.WriteString(" import=")
		w.WriteString(strconv.Quote(s.Import.Path))
		if s.Import.Default {
			w.WriteString(" default")
		}
	}
	return w.String()
}

func (s *Symbol) ToVar() *js.Var {
	return &js.Var{
		Data: []byte(s.Name),
	}
}
