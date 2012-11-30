/* 
 This program reads expressions from a file and then
 tries to unify them. In the file, sets of expressions
 are separated with a #. For example, the file

   #
   ?x1 = g(?x2)
   f(?x1,h(?x1),?x2) = f(g(?x3),?x4,?x3)
   #
   ?y = h(a(),b(),c())
   g(?y) = f(h(?y))
   h(?z) = g(?p,q()))

 has a set with two expressions and then another with
 three epxressions.

 The syntax of input expressions is constrained:

   -- variable terms must begin with a ? and contain at least
      one more character

   -- non-variable terms (that is, functional terms) must begin
      with a character other than ? (traditionally a lowercase
      letter) followed immediately by an argument list, which
      begins with a '(' and ends with a ')'. A functional
      term with no arguments is, semantically, a constant.

 The program assumes that the intput expressions are syntactically 
 correct. If the expressions are ill-formed, the program's output
 is indeterminate.

 The file name can be given as a command-line argument. If no argument 
 is provided, the file name defaults to "default.in".

 The program uses goroutines to search concurrently for a most-general
 unifier within each expression set. The program concludes with a 
 report on each expression set.
*/

package main

import (
	"fmt"
	"os"
	"io/ioutil"
	"strings"
	"sync"
	"sort"
)

//;;;

type ReasonForFailure struct {
	reason  string
	details string
}

type ExpressionSet struct {
	expressions []string
	id          int
	bindings    map[string]string
	failed      bool
	failure     ReasonForFailure
}

//;;;

func main() {
	expressionSets := getExpressionSets()
	unify(expressionSets)
	report(expressionSets)
}

func unify(expressionSets []*ExpressionSet) {
	var wg sync.WaitGroup
	for _, set := range expressionSets {
		wg.Add(1)    // add task to WaitGroup

		go func(set *ExpressionSet) {
			unifySet(set)
			wg.Done() // signal task is done
		}(set)
	}
	wg.Wait()
}

func unifySet(set *ExpressionSet) {
	// Loop through the expressions in the set.
	// splitAndTrim also canonicalizes the expression so that if
	// the expression has a standalone variable as the right term,
	// then the variable becomes the left term
	for _, expr := range set.expressions {
		if strings.Contains(expr, "=") {
			var left, right = splitAndTrim(expr)
			bindTerms(set, left, right)
		}
	}
}

func bindTerms(set *ExpressionSet, left string, right string) {
	// A term is either a variable such as ?x or a function 
   // such as f(a(), ?x, c()). A constant such as is a function 
   // of no arguments, for instance, a().
	switch {
	case isLogicalVariable(left):
		bindVariableToTerm(set, left, right)
	case isFunctionalTerm(left):
		unifyFunctionalTerms(set, left, right)
	}
}

func bindVariableToTerm(set *ExpressionSet, left string, right string) {
	if passesOccursCheck(left, right) {
		set.bindings[left] = right
	} else { 
		setReasonAndDetails(set, 
			                 "Fails occurs check", 
                          left + " occurs in " + right)
	}
}

func unifyFunctionalTerms(set *ExpressionSet, left string, right string) {
	if sameFunction(left, right) { // same name?
		same_arity, left_args, right_args := sameArity(left, right)
		if same_arity { 
			// Try to bind corresponding terms in the two expressions.
			for i := 0; i < len(left_args); i++ {
				left, right := canonicalize(string(left_args[i]), string(right_args[i]))
				bindTerms(set, left, right)
			}
		} else { // different arities
			lcount := len(left_args)
			rcount := len(right_args)
			details := fmt.Sprintf("%s: %v %s ## %s: %v %s",
			 	                    left, lcount, "args",
				                    right, rcount, "args")
			setReasonAndDetails(set, 
                             "Different arities", 
                              details)
		}
	} else { // different functions
		details := string(left[0]) + " != " + string(right[0])
		setReasonAndDetails(set, 
                          "Different functions", 
                          details)
	}
}

// Is an expression a functional term?
func isFunctionalTerm(expr string) bool {
	return !isLogicalVariable(expr)
}

// Do two functional terms have the same function?
func sameFunction(f1 string, f2 string) bool {
	return f1[0] == f2[0]
}

// Is an expression a variable (i.e., ?<alphanumeric characters>)?
func isLogicalVariable(expr string) bool {
	return string(expr[0]) == "?" && len(expr) > 1
}

// Do two functional terms have the same name?
func sameFunctionName(expr1 string, expr2 string) bool {
	return expr1 == expr2
}

// Do two functional terms have the same number of arguments?
func sameArity(left string, right string) (bool, []string, []string) {
	left_args := findArgs(left)
	right_args := findArgs(right)
	return len(left_args) == len(right_args), left_args, right_args
}

// Does a variable to be bound to an expression occur in 
// that expression?
func passesOccursCheck(v string, e string) bool {
	return !strings.Contains(e, v)
}

func report(expressionSets []*ExpressionSet) {
	for _, set := range expressionSets {
		reportAux(set)
	}
}

// Report on the expressions, before and after variables
// have been replaced by their values.
func reportAux(set *ExpressionSet) {
	switch {
	case set.failed:
		dumpReason(set)
	default:
		dumpBindings(set)
	} 
}

//;;;

/** Utilities **/

func buildExpressionSets(exps string) []*ExpressionSet {
	var expressions = splitString(exps, "\n")
	if len(expressions) < 2 {
		notifyAndMaybeDie("Need >= 2 expressions to unify.", true)
	}

   var setId = 1
	var set *ExpressionSet
	var expressionSets = []*ExpressionSet{}

	for _, expr := range expressions {
		// Create set, assign id, and add set to list.
		if strings.Contains(expr, "#") {
         set = new(ExpressionSet)
			set.id = setId
			setId++
			set.bindings = make(map[string]string)
			expressionSets = append(expressionSets, set)
		// Add expression to current set.
		} else if len(expr) > 0 {
			set.expressions = append(set.expressions, expr)
		}
	}
	return expressionSets
}

// If the form is <non-variable term> = <variable>, then
// change to <variable> = <non-variable term>
func canonicalize(left string, right string) (string, string) {
	if isLogicalVariable(right) {
		return right, left
	}
	return left, right
}

// Print bindings for successful unification
func dumpBindings(set *ExpressionSet) {
	fmt.Println()
	msg := fmt.Sprintf("%s %v", "### MGU for set", set.id)
	fmt.Println(msg)
	
	fmt.Println("\nBindings:")
	keys := []string{}
	for key := range set.bindings {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		value := set.bindings[key]
		pair := fmt.Sprintf("   %v ==> %v", key, value)
		fmt.Println(pair)
	}

	fmt.Println("\nExpressions:")
	for _, expr := range set.expressions {
		fmt.Println("   Original:  " + expr)
		fmt.Println("   Rewritten: " + substituteBindings(expr, set))
		fmt.Println()
	}
}

// Print reason for a failure to unify expressions in a set.
func dumpReason(set *ExpressionSet) {
	msg := fmt.Sprintf("%s %v", "\n### No MGU for set ", set.id)
	fmt.Println()
	fmt.Println(msg)
	fmt.Println(set.failure.reason)
	fmt.Println(set.failure.details)
}

// Extract arguments from a functional term.
func extractArgs(s string) []string {
	count := 0
	list := []string{}
	buff := ""
	for i := 0; i < len(s); i++ {
		t := string(s[i])
		switch {
		case t == "(":
			buff += t
			count++
		case t == ")":
			buff += t
			count--
		case t == ",":
			if count == 0 {
				list = append(list, buff)
				buff = ""
			} else {
				buff += t
			}
		default:
			if !strings.ContainsAny(t, " \t\n") {
				buff += t
			}
		}
	}
	list = append(list, buff) // last term
	return list
}

// Find the arguments for a functional term.
func findArgs(s string) []string {
	left := strings.Index(s, "(")
	return extractArgs(s[left + 1:len(s) - 1])
}

func getExpressionSets() []*ExpressionSet {
	expressionSets := buildExpressionSets(readInput(getFileName()))
	printExpressionSets(expressionSets)
	return expressionSets
}

func getFileName() string {
	var file_name = "default.in"
	if len(os.Args) > 1 {file_name = os.Args[1]}
	return file_name
}

func notifyAndMaybeDie(msg string, die bool) {
	fmt.Println("\n!!! " + msg);
	if die {
		os.Exit(-1)
	}
}

func printExpressionSets(expressionSets []*ExpressionSet) {
	msg := fmt.Sprintf("%v%s", len(expressionSets), " expression sets:")
	fmt.Println("\n" + msg);

	for _, set := range expressionSets {
		msg = fmt.Sprintf("%s%v", "\nSet ", set.id)
		fmt.Println(msg)
		for _, expr := range set.expressions {
			fmt.Println(expr)
		}
	}
}

func readInput(file_name string) string {
	exps, err := ioutil.ReadFile(file_name)
	if err != nil {
		notifyAndMaybeDie("Cannot read " + file_name + ". Exiting.", true)
	}
	return string(exps)
}

func splitAndTrim(str string) (string, string) {
	exprs := splitString(str, "=")
	left := strings.TrimSpace(exprs[0])
	right := strings.TrimSpace(exprs[1])
	// For an expression with a variable on one side
	// and a term on the other
	//
	//    f(a()) = ?x1
	//
	// the canonical form is
	//
	//    ?x1 = f(a())
	//
	// with the variable on the left side.
	return canonicalize(left, right)
}

func setReasonAndDetails(set *ExpressionSet, reason string, details string) {
	set.failed = true
	set.failure.reason = reason
	set.failure.details = details
}

func splitString(in string, delimiter string) []string {
	return strings.Split(in, delimiter)
}

func substituteBindings(expr string, set *ExpressionSet) string {
	expr = substituteBindingsAux(expr, set)

	// One variable bound to another? If so, a 2nd pass is in order.
	if strings.Contains(expr, "?") {
		expr = substituteBindingsAux(expr, set)
	}
	return expr
}

func substituteBindingsAux(expr string, set *ExpressionSet) string {
	for k, v := range set.bindings {
		expr = strings.Replace(expr, k, v, -1)
	}
	return expr
}
