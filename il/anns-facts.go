package atmoil

const annFactDescIndent = "  "

type iAnnFact interface {
	description(string) string
}

// collection of facts that "all must be true based on analysis",
// if they contradict we'll find out later, in here they're just all collected.
// for the `irNodeBase.facts` usage of `AnnFactAll`, the first is "the one
// indisputable core truth" about the node and others are derived from context
type AnnFactAll struct {
	Core    iAnnFact
	Derived AnnFacts
}

func (me *AnnFactAll) Description() string { return me.description("") }

func (me *AnnFactAll) description(p string) (d string) {
	// pref := p + annFactDescIndent
	return /*p + "CORE / INTRINSIC:\n" + me.Core.description(pref) + "\n" +*/ me.Derived.describe("(DERIVED) ALL OF:", p)
}

func (me *AnnFactAll) Reset() { me.Core, me.Derived = nil, nil }

// this originates from `IrIdentName`s (due to their `Anns.Candidates`)
// but may as a consequence also show up in the ancestors
type AnnFactAlts struct {
	Possibilities AnnFacts
}

func (me *AnnFactAlts) description(p string) (d string) {
	return me.Possibilities.describe("ONE OF:", p)
}

// string or uint64 or float64
type AnnFactLit struct {
	Value interface{}
	Str   func() string
}

func (me *AnnFactLit) description(p string) string { return p + "always equals: " + me.Str() }

type AnnFactDbg struct {
	Msg string
}

func (me *AnnFactDbg) description(p string) string { return p + "‹dbg›" + me.Msg }

type AnnFactTag struct {
	Value string
}

func (me *AnnFactTag) description(p string) string { return p + "always equals: " + me.Value }

// invalid node (the corresponding errs are already on record)
type AnnFactUndef struct{}

func (me *AnnFactUndef) description(p string) string { return p + "undefined / unreachable / invalid" }

// reference: all facts for `To` are my facts
type AnnFactRef struct {
	To INode
}

func (me *AnnFactRef) description(p string) string {
	return p + "REF:\n" + me.To.Facts().description(p+annFactDescIndent)
}

type AnnFactCall struct {
	Callee *AnnFactRef
	Arg    *AnnFactRef
}

func (me *AnnFactCall) description(p string) (d string) {
	pref := p + annFactDescIndent
	d = p + "call:\n" + me.Callee.description(pref) + "\n" + p + "with:\n" + me.Arg.description(pref)
	return
}

type AnnFactCallable struct {
	Arg *AnnFactRef
	Ret *AnnFactRef
}

func (me *AnnFactCallable) description(p string) (d string) {
	pref := p + annFactDescIndent
	d = p + "callable: returns\n" + me.Ret.description(pref) + "\n" + p + "given arg:\n" + me.Arg.description(pref)
	return
}

type AnnFacts []iAnnFact

func (me *AnnFacts) Add(facts ...iAnnFact) {
	// *me = append(*me, facts...)
	// return
	this := *me
	var merge bool
	for _, f := range facts {
		if _, merge = f.(AnnFacts); merge {
			break
		}
	}
	if !merge {
		this = append(this, facts...)
	} else {
		for _, f := range facts {
			if fs, ok := f.(AnnFacts); ok {
				this = append(this, fs...)
			} else {
				this = append(this, f)
			}
		}
	}
	*me = this
}

func (me AnnFacts) description(p string) string {
	return me.describe("MANY:", p)
}

func (me AnnFacts) describe(label string, p string) (d string) {
	d = p + label
	pref := p + "->"
	for i := range me {
		d += "\n" + me[i].description(pref)
	}
	return
}
