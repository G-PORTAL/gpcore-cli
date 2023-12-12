package generator

import (
	. "github.com/dave/jennifer/jen"
	"github.com/stoewer/go-strcase"
	"strings"
)

var (
	apiClientImport string
	apiGRPCImport   string
)

// GenerateSubCommand generates a subcommand based on the given metadata. The
// targetFilename is the file where the generated code will be saved to.
func GenerateSubCommand(metadata SubcommandMetadata, targetFilename string) error {
	packageName := strings.Replace(strings.Replace(metadata.Definition.Name, ".", "_", -1), "-", "_", -1)
	f := NewFile(packageName)
	warningComment(f)

	name := strcase.LowerCamelCase(metadata.Name)

	// Imports
	apiClientImport = "buf.build/gen/go/gportal/gpcore/protocolbuffers/go/gpcore/api/" + metadata.Action.APICall.Client + "/" + metadata.Action.APICall.Version
	apiGRPCImport = "buf.build/gen/go/gportal/gpcore/grpc/go/gpcore/api/" + metadata.Action.APICall.Client + "/" + metadata.Action.APICall.Version + "/" + metadata.Action.APICall.Client + metadata.Action.APICall.Version + "grpc"

	f.ImportName("github.com/spf13/cobra", "cobra")
	f.ImportName("github.com/G-PORTAL/gpcore-cli/pkg/client", "client")
	f.ImportName("github.com/G-PORTAL/gpcore-cli/pkg/config", "config")
	f.ImportName("google.golang.org/grpc", "grpc")
	f.ImportName("github.com/charmbracelet/ssh", "ssh")
	f.ImportName("github.com/jedib0t/go-pretty/v6/table", "table")
	f.ImportAlias(apiClientImport, apiClient(metadata))
	f.ImportName(apiGRPCImport, apiClient(metadata)+"grpc")

	// Parameters (variables)
	for _, param := range metadata.Action.Params {
		f.Var().Add(variableDefinition(name, param))
	}
	f.Line()

	// Enum helper functions
	for _, param := range metadata.Action.Params {
		if enumType(param.Type) {
			f.Add(enumToProtoFunc(param.Type)...)
		}
	}

	// Build up the command
	values := Dict{
		Id("Use"):                   Lit(metadata.Name),
		Id("Short"):                 Lit(metadata.Action.Description),
		Id("Long"):                  Lit(metadata.Action.Description),
		Id("DisableFlagsInUseLine"): True(),
		Id("Args"):                  Qual("github.com/spf13/cobra", "OnlyValidArgs"),
		Id("RunE"): Func().Params(
			Id("cobraCmd").Op("*").Qual("github.com/spf13/cobra", "Command"),
			Id("args").Index().String()).Error().
			Block(runCommand(name, metadata)...),
	}

	// Add flags and params
	if len(metadata.Action.Params) > 0 {
		var args []Code
		for _, v := range metadata.Action.Params {
			args = append(args, Lit(strcase.KebabCase(v.Name)))
		}
		values[Id("ValidArgs")] = Index().String().Values(args...)
	}

	// Final command
	f.Var().Add(Id(name+"Cmd").Op("=").
		Op("&").Qual("github.com/spf13/cobra", "Command").
		Values(values))

	f.Func().Id("init").Params().Block(initFunc(name, metadata)...)

	return f.Save(targetFilename)
}

// runCommand generates the code for the RunE function of the command. This
// function will call the API and print the response.
func runCommand(name string, metadata SubcommandMetadata) []Code {
	c := make([]Code, 0)

	// Context
	c = append(c, Id("ctx").
		Op(":=").
		Qual("github.com/G-PORTAL/gpcore-cli/pkg/client", "ExtractContext").Call(
		Id("cobraCmd")))

	// Identifier we use to make queries to that API. This identifier (an ID)
	// is used as the "Id" field in the request. This is useful for actions
	// which operate on a specific resource. The identifier itself can be
	// a pointer to something from the session (like the current project ID)
	// or nil, if the identifier is not required (list actions).
	// Identifiers can be set on the action or the definition. If the action
	// is set, the action identifier is used. If not, the definition identifier
	// is used.
	identifier := ""
	// "Global identifier for all actions present?
	if metadata.Definition.Identifier != "" {
		identifier = metadata.Definition.Identifier
	}
	// Override with action identifier
	if metadata.Action.Identifier != "" {
		// Reset identifier on purpose (override) global identifier
		if metadata.Action.Identifier == "nil" {
			identifier = ""
		} else {
			// Action specific identifier
			identifier = metadata.Action.Identifier
		}
	}

	if identifier != "" {
		c = append(c, Id("session").Op(":=").
			Id("ctx").
			Dot("Value").Call(Lit("config")).
			Assert(Op("*").Qual("github.com/G-PORTAL/gpcore-cli/pkg/config", "SessionConfig")))
		c = append(c, If(Id("session").Op("==").Nil()).Block(
			Return(Qual("fmt", "Errorf").Call(Lit("no session found, please login first")))))
		c = append(c, If(Id(identifier).Op("==").Nil()).Block(
			Return(Qual("fmt", "Errorf").Call(Lit("no identifier found, please set the identifier first")))))
		c = append(c, Line())
	}

	// Pre hook
	if hasHook(metadata.Definition.Name, name, "pre") {
		c = append(c, addHook(name, "pre")...)
		c = append(c, Line())
	}

	// API call
	c = append(c, Id("grpcConn").Op(":=").
		Id("ctx").
		Dot("Value").Call(Lit("conn")).
		Assert(Op("*").Qual("google.golang.org/grpc", "ClientConn")))
	c = append(c, Id("client").Op(":=").
		Qual(apiGRPCImport, "New"+title(metadata.Action.APICall.Client)+"ServiceClient").Call(Id("grpcConn")))

	apiCallParams := Dict{}
	// Do we have an identifier?
	if identifier != "" {
		apiCallParams[Id("Id")] = Op("*").Id(identifier)
	}

	// Specific parameters set?
	for _, param := range metadata.Action.Params {
		variable := strcase.LowerCamelCase(name) + title(strcase.LowerCamelCase(param.Name))
		var val *Statement
		if enumType(param.Type) {
			// Enum helper function call
			val = Id(stripPackage(param.Type) + "ToProto").Call(Id(variable))
		} else {
			// Optional pointer type
			if !param.Required && param.Default == nil {
				val = Op("&").Id(variable)
			} else {
				val = Id(variable)
			}
		}
		apiCallParams[Id(title(strcase.LowerCamelCase(param.Name)))] = val
	}

	c = append(c, List(Id("resp"), Id("err")).Op(":=").
		Id("client").Dot(metadata.Action.APICall.Endpoint).Call(
		Id("cobraCmd").Dot("Context").Call(),
		Op("&").Qual(apiClientImport, metadata.Action.APICall.Endpoint+"Request").Values(apiCallParams)))

	c = append(c, If(Id("err").Op("!=").Nil()).Block(
		Return(Id("err"))))

	if strings.HasPrefix(metadata.Action.APICall.Endpoint, "List") {
		// List response
		c = append(c, listResponse(metadata)...)
	} else {
		// Single response
		c = append(c, singleResponse(metadata)...)
	}

	c = append(c, Line())

	c = append(c, If(Qual("github.com/G-PORTAL/gpcore-cli/pkg/config", "JSONOutput")).Block(
		List(Id("jsonData"), Id("err")).
			Op(":=").
			Qual("encoding/json", "MarshalIndent").Call(
			Id("respData"),
			Lit(""),
			Lit("  ")),
		If(Id("err").Op("!=").Nil()).Block(
			Return(Id("err"))),
		Id("cobraCmd").Dot("Println").Call(
			Id("string").Call(
				Id("jsonData")))))

	c = append(c, Return(Nil()))

	return c
}

// listResponse generates the code for the list response. This function will
// print the response as a table.
func listResponse(metadata SubcommandMetadata) []Code {
	c := make([]Code, 0)

	c = append(c, Id("sshSession").Op(":=").
		Id("ctx").Dot("Value").
		Call(Lit("ssh")).
		Assert(Op("*").Qual("github.com/charmbracelet/ssh", "Session")))

	// Remove root key
	rootKey := metadata.Action.RootKey
	if rootKey == "" {
		rootKey = title(strcase.LowerCamelCase(plural(metadata.Name)))
	}
	c = append(c, Id("respData").Op(":=").
		Id("resp").Dot(rootKey))

	// Build the table
	c = append(c, Id("headerRow").Op(":=").
		Qual("github.com/jedib0t/go-pretty/v6/table", "Row").Values())
	c = append(c, Id("tbl").Op(":=").
		Qual("github.com/jedib0t/go-pretty/v6/table", "NewWriter").Call())
	c = append(c, Id("tbl").Dot("SetStyle").
		Call(Id("table").Dot("StyleRounded")))
	c = append(c, Id("tbl").Dot("SetOutputMirror").
		Call(Op("*").Id("sshSession")))
	c = append(c, Id("cobraCmd").Dot("SetOut").
		Call(Op("*").Id("sshSession")))

	// Call post hook if available
	if hasHook(metadata.Definition.Name, metadata.Name, "post") {
		c = append(c, List(Id("respHook"), Id("err")).Op(":=").
			Id(title(strcase.LowerCamelCase(metadata.Name))+"HookPost").
			Call(Id("resp"), Id("cobraCmd")))
		c = append(c, If(Id("err").Op("!=").Nil()).Block(
			Return(Id("err"))))
	}

	// Collect rows
	if len(metadata.Action.Fields) == 0 {
		// We use all the fields
		c = append(c, For(List(Id("j"), Id("entry")).
			Op(":=").Range().Id("respData")).Block(
			Id("c").Op(":=").
				Qual("reflect", "ValueOf").
				Call(Op("*").Id("entry")),
			Id("row").Op(":=").
				Qual("github.com/jedib0t/go-pretty/v6/table", "Row").Values(),
			For(
				Id("i").Op(":=").Lit(0),
				Id("i").Op("<").Id("c").Dot("NumField").Call(),
				Id("i").Op("++")).Block(
				If(Id("c").Dot("Type").Call().Dot("Field").Call(Id("i")).Dot("IsExported").Call().Block(
					Id("val").Op(":=").Qual("fmt", "Sprintf").Call(
						Lit("%v"),
						Id("c").Dot("Field").Call(
							Id("i")).Dot("Interface").Call()),
					Id("col").Op(":=").Qual("fmt", "Sprintf").Call(
						Lit("%v"),
						Id("c").Dot("Type").Call().
							Dot("Field").Call(
							Id("i")).
							Dot("Name")),
					Switch(Id("col").Block(
						Case(Lit("CreatedAt")).Block(tableColValue("CreatedAt")...),
						Case(Lit("Currency")).Block(tableColValue("Currency")...),
						Case(Lit("Environment")).Block(tableColValue("Environment")...),
						Id("row").Op("=").Append(Id("row"), Id("val")))),
					If(Id("j").Op("==").Lit(0)).Block(
						Id("headerRow").Op("=").Append(
							Id("headerRow"),
							Id("c").Dot("Type").Call().Dot("Field").Call(Id("i")).Dot("Name"))),
				))),
			Id("tbl").Dot("AppendRow").Call(Id("row"))))

		// Add the header row
		c = append(c, Id("tbl").Dot("AppendHeader").Call(Id("headerRow")))

	} else {
		// Header row
		for _, value := range metadata.Action.Fields {
			c = append(c, Id("headerRow").Op("=").Append(Id("headerRow"), Lit(value)))
		}
		c = append(c, Id("tbl").Dot("AppendHeader").Call(Id("headerRow")))
		c = append(c, Line())

		// Only use whitelisted fields
		valuesCode := make([]Code, 0)
		indexVariable := "_"

		valuesCode = append(valuesCode, Id("row").Op(":=").Make(Index().Interface(), Lit(0)))
		valuesCode = append(valuesCode, Id("val").Op(":=").Lit(""))

		// Rows
		for _, value := range metadata.Action.Fields {
			// Default format
			valuesCode = append(valuesCode, Id("val").Op("=").
				Qual("fmt", "Sprintf").Call(
				Lit("%v"),
				Id("entry").Dot(value)))

			// Call hook
			if hasHook(metadata.Definition.Name, metadata.Name, "post") {
				indexVariable = "i"
				valuesCode = append(valuesCode, If(List(Id("v"), Id("ok")).Op(":=").
					Id("respHook").Index(Id("i")).Index(Lit(title(value))).Op(";").Id("ok").Block(
					Id("val").Op("=").Id("v"))))
			}

			// Special formatting
			valuesCode = append(valuesCode, tableColValue(value)...)

			// Append to row
			valuesCode = append(valuesCode, Id("row").Op("=").Append(Id("row"), Id("val")))
		}

		valuesCode = append(valuesCode, Id("tbl").Dot("AppendRow").Call(Id("row")))
		c = append(c, For(List(Id(indexVariable), Id("entry")).Op(":=").Range().Id("respData")).Block(
			valuesCode...))

	}

	// Do we have CSV output?
	c = append(c, Line())
	c = append(c, If(
		Qual("github.com/G-PORTAL/gpcore-cli/pkg/config", "CSVOutput")).Block(
		Id("tbl").Dot("RenderCSV").Call(),
		Return(Nil())))

	// Normal output
	c = append(c, If(
		Op("!").Qual("github.com/G-PORTAL/gpcore-cli/pkg/config", "JSONOutput")).Block(
		Id("tbl").Dot("Render").Call()))

	return c
}

// singleResponse generates the code for the single response.
func singleResponse(metadata SubcommandMetadata) []Code {
	c := make([]Code, 0)

	c = append(c, Id("respData").Op(":=").Id("resp"))

	return c
}

// tableColValue generates the code for special table column values. Some
// columns need special formatting, like the "CreatedAt" column, which is a
// timestamp.
func tableColValue(col string) []Code {
	c := make([]Code, 0)

	switch col {
	case "CreatedAt":
		// This one is tricky. We need to get the seconds from the struct
		// and convert it to a time.Time object. Problem is the Go type system
		// and the fact that we don't know the type of the field. So we need to
		// use reflection to get the value and then convert it to a time.Time.
		// A lib like Jennifer would be better suited for this.
		c = append(c, Id("c").Op(":=").Qual("reflect", "ValueOf").Call(
			Op("*").Id("entry")))
		c = append(c, Id("a").Op(":=").
			Qual("reflect", "Indirect").Call(
			Id("c").Dot("FieldByName").Call(Lit("CreatedAt"))))
		c = append(c, Id("s").Op(":=").
			Qual("reflect", "Indirect").Call(
			Id("a")).Dot("FieldByName").Call(Lit("Seconds")))
		c = append(c, If(Id("s").Dot("CanInt").Call().Block(
			Id("t").Op(":=").Qual("time", "Unix").Call(
				Id("s").Dot("Int").Call(),
				Lit(0)),
			Id("val").Op("=").
				Id("t").Dot("Format").Call(Lit("2006-01-02 15:04:05")))))
	case "Currency":
		c = append(c, Id("val").Op("=").
			Id("val").Index(Lit(9), Empty()))
	case "Environment":
		c = append(c, Id("val").Op("=").
			Id("val").Index(Lit(20), Empty()))
	}

	return c
}

// initFunc generates the init function, which will add all flags and params to
// the command and add it to the root command. Required flags are marked as
// such.
func initFunc(name string, metadata SubcommandMetadata) []Code {
	c := make([]Code, 0)

	// Params
	for _, param := range metadata.Action.Params {
		dataType := title(param.Type)
		if enumType(param.Type) {
			dataType = "String"
		}

		c = append(c,
			Id(strcase.LowerCamelCase(name)+"Cmd").
				Dot("Flags").Call().
				Dot(dataType+"Var").Params(
				Op("&").Id(strcase.LowerCamelCase(name)+title(strcase.LowerCamelCase(param.Name))),
				Lit(strcase.KebabCase(param.Name)),
				defaultValue(param),
				Lit(parameterDescription(param))))
	}

	c = append(c, Line())

	// Required fields
	for _, param := range metadata.Action.Params {
		if param.Required {
			c = append(c,
				Id(strcase.LowerCamelCase(name)+"Cmd").
					Dot("MarkFlagRequired").
					Call(Lit(strcase.KebabCase(param.Name))))
		}
	}

	c = append(c, Line())

	// Add the command to the root command
	c = append(c, Id("Root"+strcase.UpperCamelCase(metadata.Definition.Name)+"Command").
		Dot("AddCommand").
		Call(Id(name+"Cmd")))

	return c
}

// enumToProtoFunc generates a function which converts a string to a proto enum
// to be used it in the API call.
func enumToProtoFunc(enumType string) []Code {
	c := make([]Code, 0)

	// Generate the function
	c = append(c, Func().Id(stripPackage(enumType)+"ToProto").Params(
		Id("a").String()).Id(enumType).Block(
		For(List(Id("k"), Id("v")).Op(":=").Range().Id(enumType+"_name").Block(
			If(Id("v").Op("==").
				Lit(strings.ToUpper(strcase.SnakeCase(stripPackage(enumType)))+"_").
				Op("+").
				Qual("strings", "ToUpper").Call(Id("a"))).Block(
				Return(Id(enumType).Call(Id("k"))),
			),
		),
			Return(Id(enumToProtoType(enumType, "UNSPECIFIED"))),
		),
	),
	)

	return c
}

// variableDefinition returns a statement for a variable definition, coming from
// a given param. We can not use the string representation of a type, so we need
// to transpile it to the correct type. Enum types are always strings.
func variableDefinition(name string, param Param) *Statement {
	variable := Id(strcase.LowerCamelCase(name) + title(strcase.LowerCamelCase(param.Name)))
	switch param.Type {
	case "bool":
		variable.Bool()
	case "int":
		variable.Int()
	case "float":
		variable.Float64()
	case "string":
		variable.String()
	default:
		if enumType(param.Type) {
			variable.String()
		} else {
			variable.Id(title(param.Type))
		}
	}
	return variable
}
