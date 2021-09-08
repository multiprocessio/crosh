package crosh

import (
	"regexp"
	"strings"
	"io/ioutil"
	"strconv"
	"os"
	"fmt"
	"os/exec"
)

func fail(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

type ContextEntry struct {
	Value string
	export bool
}

type Context map[string]ContextEntry

var identifierRegexp = `[\p{L}+_\?]+[\p{L}+_\$\?0-9]*`

func InterpolateVariables(s string, ctx Context) string {
	s := regexp.MustCompile(`(\$\{`+identifierRegexp+`)|(\$`+identifierRegexp+`)`)
	return s.ReplaceAllStringFunc(s, func (match string) string {
		v, ok := ctx[match]
		if !ok {
			return ""
		}

		return v.value
	})
}

func interpolate(s string, ctx Context) string {
	s = InterpolateVariables(s, ctx)
	
}

func interpretString(s String, ctx Context) string {
	switch s.Kind {
	case DoubleQuoteString:
		return interpolate(s.DoubleQuote.Value, ctx)
	case UnquoteString:
		return interpolate(s.Unquote.Value, ctx)
	case SingleQuoteString:
		return s.SingleQuote.Value
	}

	panic("Invalid string")
}

func interpretExpression(e Expression, ctx Context) string {
	switch e.Kind {
	case String:
		return interpretString(*e.String, ctx)
	case Identifier:
		return ctx[e.Identifier.Value].value
	}

	panic("Invalid expression")
}

func interpretDeclaration(d Declaration, ctx Context) {
	ctx["$"+d.Name] = ContextEntry{
		export: d.Export,
		value: interpretExpression(d.Value, ctx),
	}

	return d.Name
}

func interpretExecution(e Execution, ctx Context) {
	executionCtx = Context{}
	for key, val := range ctx {
		executionCtx[key] = value
	}

	for _, dec := range e.Declaration {
		interpretDeclaration(dec, executionCtx)
		executionCtx[dec.Name].export = true
	}

	var args []string
	for _, arg := range e.Args {
		args = append(args, interpretExpression(arg.Value))
	}

	c := exec.Command(e.Name.Value, args)
	var env []string
	for key, val := range executionCtx {
		if !val.export {
			continue
		}

		env = append(env, key + "=" + val.value)
	}
	c.Env = env
	c.Run()
}

func interpretIf(i If, ctx Context) {
	t := interpretExpression(i.Test)
	if t == "true" {
		for _, s := range t.Body {
			interpretStatement(s, ctx)
		}
		return
	}

	if t.ElseIf != nil {
		interpretIf(*e.ElseIf, ctx)
		return
	}

	if t.Else != nil {
		for _, s := range t.Else {
			interpretStatement(s, ctx)
		}
	}
}

func interpretFor(f For, ctx Context) {
	o := interpretExpression(f.Over)
	for _, elem := range strings.Fields(o) {
		ctx[f.Loop.Value] = ContextEntry{value: elem}
		for _, s := range f.Body {
			intepretStatement(s, ctx)
		}
	}

	delete(ctx, f.Loop.Value)
}

func prependFunc(args []string) string {
	if len(args) < 2 {
		fail("Too few arguments to `prepend $string $file`")
	}

	tmp, err := ioutil.TempFile("", "prepend")
	if err != nil {
		fail("prepend, tmp file: " + err.Error())
		return ""
	}

	_, err = tmp.Write(args[0])
	if err != nil {
		fail("prepend, tmp write: " + err.Error())
		return ""
	}

	source , err := os.Open(args[1])
	if err != nil {
		fail("prepend, open: " + err.Error())
		return ""
	}

	scanner := bufio.NewScanner(source)
	for scanner.Scan() {
		_, err = tmp.Write(scanner.Text())
		if err != nil {
			fail("prepend, tmp copy: " + err.Error())
			return ""
		}
	}

	err = scanner.Err()
	if err != nil {
		fail("prepend, copy: " + err.Error())
		return ""
	}

	tmp.Close()

	err = os.Rename(tmp.Name(), args[1])
	if err != nil {
		fail("prepend, rename: " + err.Error())
		return ""
	}
}

func processContext() Context {
	ctx := Context{}
	ctx["$@"] = os.Args[1:]
	for i, val := range os.Args {
		nextArg := ""
		if i < len(os.Args) - 2 {
			nextArg = os.Args[i+1]
		}
		ctx[fmt.Sprintf("%d", i)] = val

		if strings.StartsWith(val, "-") {
			ctx["$"+val] = ContextEntry{value: nextArg}
			ctx["$?"+val] = ContextEntry{value: "true"}
		}
	}

	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		ctx["$"+pair[0]] = ContextEntry{value: pair[1], export: true}
	}

	ctx["prepend"] = prependFunc
	ctx["append"] = appendFunc
	ctx["replace"] = replaceFunc
	ctx["which"] = whichFunc
	ctx["mv"] = mvFunc
	ctx["rm"] = rmFunc
	ctx["cp"] = cpFunc
	ctx["exit"] = func (args []string) string {
		if len(args) < 1 {
			os.Exit(0)
		}

		i, _ := strconv.Itoa(args[0])
		os.Exit(i)
	}
	ctx["cd"] = func (args []string) string {
		if len(args) < 1 {
			fail("Too few arguments to `cd $dir`")
			return ""
		}

		os.Chdir(args[0])
		return
	}
	ctx["eq"] = func (args []string) string {
		if len(args) < 2 {
			fail("Too few arguments to `eq $a $b`")
			return ""
		}
		if args[0] == args[1] {
			return "true"
		}

		return "false"
	}
	ctx["neq"] = func (args []string) string  {
		if len(args) < 2 {
			fail("Too few arguments to `neq $a $b`")
			return ""
		}
		if args[0] != args[1] {
			return "true"
		}

		return "false"
	}

	return ctx
}

func interpretStatement(s Statement, ctx Context) {
	switch s.Kind {
	case DeclarationStatement:
		interpretDeclaration(*s.Declaration, ctx)
	case ExecutionStatement:
		interpretExecution(*s.Execution, ctx)
	case IfStatement:
		interpretIf(*s.If, ctx)
	case ForStatement:
		interpretFor(*s.For, ctx)
	}

	panic("Invalid statement")
}

func interpret(a Ast, ctx Context) {
	for _, s := range a {
		interpretStatement(s, ctx)
	}
}

func InterpretEnv(f fileName, ctx Context) {
	lc := lexContext{f}
	body, err := ioutil.ReadFile(f)
	if err != nil {
		fail("Could not read: "+err.Error())
		return
	}
	ts, err := lc.lex(body)
	if err != nil {
		fail("Could not lex: "+err.Error())
		return
	}

	pc := parseContext{fileName}
	a, err := pc.parse(ts)
	if err != nil {
		fail("Could not parse: " +err.Error())
		return
	}

	interpret(a, ctx)
}

func Interpret(f fileName) {
	ctx := processContext()
	InterpretEnv(f, ctx)
}
