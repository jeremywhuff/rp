package main

import (
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"regexp"
	"strings"
	"text/template"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	. "github.com/jeremywhuff/rp"
)

// Thoughts on next steps:
// - Create a function that uses the ast package to break a large amount of source code into sections, breaking at the
//       if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return } statements.
// - Create a function that uses the ast package to find undeclared variable names for these subsections of code.
// - Create a function that loops through each subsection starting with the last one. For each subsection, it should
//       add c.Get() calls at the top for each undeclared variable name and then append the variable name to a slice of
//       variable names that need a matching c.Set call to be created in another subsection.
// - Create a function that uses the ast package to detect if a given variable name is assigned for these subsections
//       of code and, if so, returns the line number of the last assignment statement.
// - Add this function into the looping function to find the last assignment statement in the last subsection that
//       contains an assignment of one of the variables in the slice of variable names that need a matching c.Set call.
//       Then, add a c.Set call immediately after that line within that subsection and remove it from the slice.
// - Also in the looping function, replace the if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error":
//       err.Error()}); return } calls with if err != nil { return err } calls.
// - When each of the subsections has all of the modifications, it should call MigrateToRP to create a chain
//       declaration of the form chainName := MakeChain(S(...)).
// - Create a command line tool that wraps all of this functionality so that users can convert sections of their code.
//       It should first request that the user copy/paste their code (unwrapped so that all of the breakpoints are
//       the root scope) and hit enter. Then, it should print out the modified code and ask the user if they want to
//       export to a file, copy to clipboard, or exit. Allow them to both operations in sequence.

func main() {

	src := `package main

	import (
		"context"
		"errors"
		"fmt"
		"log"
		"net/http"
		"time"
	
		"github.com/gin-gonic/gin"
		. "github.com/jeremywhuff/rp"
		"github.com/jeremywhuff/rp/rpout"
		"go.mongodb.org/mongo-driver/bson/primitive"
		"go.mongodb.org/mongo-driver/mongo"
		"go.mongodb.org/mongo-driver/mongo/options"
	)

	func main() {
	
	// Parse request body

	var body PurchaseRequestBody
	err := c.ShouldBindJSON(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch customer

	customer := CustomerDocument{}
	err = mongoClient.Database("rp_test").Collection("customers").FindOne(context.Background(),
		map[string]any{
			"customer_id": body.CustomerID,
		}).Decode(&customer)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Fetch inventory

	item := InventoryDocument{}
	err = mongoClient.Database("rp_test").Collection("inventory").FindOne(context.Background(),
		map[string]any{
			"sku": body.SKU,
		}).Decode(&item)
	if err == nil {
		log.Println("Hello")
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check inventory stock

	if item.Stock < body.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Not enough stock"})
		return
	}
	}`

	sections, err := BreakIntoSections(src)
	if err != nil {
		panic(err)
	}

	log.Printf("Found %d sections", len(sections))

	for _, section := range sections {
		log.Println(section)
		log.Println("**********")
	}
}

//   - Create a function that uses the ast package to break a large amount of source code into sections, breaking at the
//     if err != nil { c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()}); return } statements.
func BreakIntoSections(src string) ([]string, error) {

	// Parse the source code
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		return nil, err
	}

	// Find all the if err != nil { ... } statements
	var sections []string
	var pos token.Pos = 0
	ast.Inspect(f, func(n ast.Node) bool {

		if stmt, ok := n.(*ast.FuncDecl); ok {
			if stmt.Name.Name == "main" {
				if stmt.Body != nil {
					log.Print("Found main function")
					sections = append(sections, src[pos:stmt.Body.Pos()])
					pos = stmt.Body.Pos() + 1
				}
			}
		}

		if stmt, ok := n.(*ast.IfStmt); ok {
			sect, p, err := findcJSON(src, pos, stmt)
			if err == nil {
				sections = append(sections, sect)
				pos = p
			}
		}
		return true
	})

	return sections, nil
}

func findcJSON(src string, pos token.Pos, stmt *ast.IfStmt) (string, token.Pos, error) {
	var str string
	var next token.Pos
	ast.Inspect(stmt, func(n ast.Node) bool {
		if call, ok := n.(*ast.CallExpr); ok {
			if exp, ok := call.Fun.(*ast.SelectorExpr); ok {
				if x, ok := exp.X.(*ast.Ident); ok {
					if x.Name == "c" && exp.Sel != nil && exp.Sel.Name == "JSON" {
						if stmt.Else != nil {
							log.Printf("Found c.JSON() at %d", stmt.Else.End())
							str = src[pos : stmt.Else.End()-1]
							next = stmt.Else.End()
						} else {
							log.Printf("Found c.JSON() at %d", stmt.Body.End())
							str = src[pos : stmt.Body.End()-1]
							next = stmt.Body.End()
						}
					}
				}
			}
		}
		return true
	})
	if str == "" {
		return "", 0, errors.New("no c.JSON() found")
	}
	return str, next, nil
}

// This function migrates source code from the style in PurchaseHandler to the style in PurchaseHandlerDirectMigrationToRP.
// For example, it converts this:
//
//	// Fetch customer
//	customer := CustomerDocument{}
//	err = mongoClient.Database("rp_test").Collection("customers").FindOne(context.Background(),
//		map[string]any{
//			"customer_id": body.CustomerID,
//		}).Decode(&customer)
//	if err != nil {
//		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//		return
//	}
//
// to this:
//
//	fetchCustomer := MakeChain(S(
//		FuncStr("fetch_customer", "req.body")+CtxOutStr("mongo.document.customer"),
//		func(in any, c *gin.Context, lgr Logger) (any, error) {
//
//			body := c.MustGet("req.body").(*PurchaseRequestBody)
//
//			customer := CustomerDocument{}
//			err := mongoClient.Database("rp_test").Collection("customers").FindOne(context.Background(),
//				map[string]any{
//					"customer_id": body.CustomerID,
//				}).Decode(&customer)
//			if err != nil {
//				return nil, err
//			}
//
//			c.Set("mongo.document.customer", &customer)
//			return nil, nil
//		}))
func MigrateToRP(src string, ctxDependencies []string, ctxOutputs []string, ctxOutputVars []string) (string, error) {

	if len(ctxOutputs) != len(ctxOutputVars) {
		return "", errors.New("ctxOutputs and ctxOutputVars must have the same length")
	}

	// Define a template for the migration
	tmpl := `{{.StageName}} := MakeChain(S(
		{{.FuncStr}}+{{.CtxOutStr}},
		func(in any, c *gin.Context, lgr Logger) (any, error) {

			{{.Body}}

			{{.SetCtx}}

			return nil, nil
		}))`

	// Define a struct to hold the template data
	type TemplateData struct {
		StageName string
		FuncStr   string
		CtxOutStr string
		Body      string
		SetCtx    string
	}

	// Fill in each of the template's fields

	firstCommentLine := ""
	firstCommentLineNumber := -1
	regex := regexp.MustCompile("[^a-zA-Z0-9_ ]+")
	for i, line := range strings.Split(src, "\n") {
		if strings.HasPrefix(line, "//") && regex.ReplaceAllString(line, "") != "" {
			firstCommentLine = strings.TrimSpace(line[2:])
			firstCommentLineNumber = i
			break
		}
	}
	srcWithoutFirstCommentLine := ""
	if firstCommentLine == "" {
		firstCommentLine = "unnamed stage"
		srcWithoutFirstCommentLine = src
	} else {
		for i, line := range strings.Split(src, "\n") {
			if i != firstCommentLineNumber {
				srcWithoutFirstCommentLine += line + "\n"
			}
		}
	}

	setCtx := ""
	for i, out := range ctxOutputs {
		setCtx += "c.Set(\"" + out + "\", " + ctxOutputVars[i] + ")"
		if i < len(ctxOutputs)-1 {
			setCtx += "\n"
		}
	}

	data := TemplateData{
		StageName: camelCase(firstCommentLine),
		FuncStr:   FuncStr(funcNameFormat(firstCommentLine), ctxDependencies...),
		CtxOutStr: CtxOutStr(ctxOutputs...),
		Body:      srcWithoutFirstCommentLine,
		SetCtx:    setCtx,
	}

	// Execute the template
	t, err := template.New("migration").Parse(tmpl)
	if err != nil {
		return "", err
	}
	var out strings.Builder
	err = t.Execute(&out, data)
	if err != nil {
		return "", err
	}

	return out.String(), nil
}

func camelCase(s string) string {

	// Remove all characters that are not alphanumeric or spaces or underscores
	s = regexp.MustCompile("[^a-zA-Z0-9_ ]+").ReplaceAllString(s, "")

	// Replace all underscores with spaces
	s = strings.ReplaceAll(s, "_", " ")

	// Title case s
	s = cases.Title(language.AmericanEnglish, cases.NoLower).String(s)

	// Remove all spaces
	s = strings.ReplaceAll(s, " ", "")

	// Lowercase the first letter
	if len(s) > 0 {
		s = strings.ToLower(s[:1]) + s[1:]
	}

	return s
}

func funcNameFormat(s string) string {

	// Remove all characters that are not alphanumeric or spaces or underscores
	s = regexp.MustCompile("[^a-zA-Z0-9_ ]+").ReplaceAllString(s, "")

	// Replace all underscores with spaces
	s = strings.ReplaceAll(s, "_", " ")

	// Convert all alphabetic characters to lowercase
	s = strings.ToLower(s)

	out := ""
	for _, word := range strings.Split(s, " ") {
		out += word
	}

	return out
}
