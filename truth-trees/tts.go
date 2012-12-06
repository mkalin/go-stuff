package main

import (
	"fmt"
)

// connectives and quantifiers
const (
   not        = "~"
	or         = "v"
	and        = "."
	implies    = "=>"
	AtLeastOne = "E"
	Any        = "A"
)

// inference and rewrite rules
type Rule struct {
	// inference rules
	const MP = "modus ponens"              // (p . (p => q)) => p
	const MT = "modus tollens"             // (~q . (p => q)) => ~p
	const CI = "conjunction introduction"  // p . q => (p . q)
	const CE = "conjunction elimination"   // (p . q) => p; (p . q) => q
	const DI = "disjunction introduction"  // p => (p v q)

	// rewrite rules
	const ME = "material equivalence"      // (p == q) => ((p => q) . (q => p))
	const DN = "double negation"           // ~~p == p
   const MI = "material implication"      // (p => q) == (~p v q)
   const DA = "disjunctive association"   // ((p v q) v r) == (p v (q v r))
   const CA = "conjunctive association"   // ((p . q) . r) == (p . (q . r))
   const DC = "disjunctive commutation"   // (p v q) == (q v p)
   const CC = "conjunctive commutation"   // (p . q) == (q . p)
	const DD = "disjunctive distribution"  // (p v (q . r)) == ((p v q) . (p v r))
	const CD = "conjunctive distribution"  // (p . (q v r)) == ((p . q) v (p . r))
}

type Premises   []string
type Conclusion string 
type ProofStep struct {
	from []Premises
	to   Conclusion
	rule 
}

type Proof struct {
	premises   Premises
	conclusion Conclusion
}


type TruthTree struct {
	name     string
	parent   *TruthTree
	children []*TruthTree
}


func main() {
	fmt.Println()
}