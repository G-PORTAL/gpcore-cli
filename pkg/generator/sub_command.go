package generator

import (
	"github.com/G-PORTAL/gpcore-cli/pkg/api"
	. "github.com/dave/jennifer/jen"
	"github.com/stoewer/go-strcase"
	"strings"
)

var (
	apiClientImport string
	apiGRPCImport   string
	apiTypesImport  string
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
	apiTypesImport = "buf.build/gen/go/gportal/gpcore/protocolbuffers/go/gpcore/type/v1"

	f.ImportName("github.com/spf13/cobra", "cobra")
	f.ImportName("github.com/G-PORTAL/gpcore-cli/pkg/client", "client")
	f.ImportName("github.com/G-PORTAL/gpcore-cli/pkg/config", "config")
	f.ImportName("github.com/G-PORTAL/gpcore-cli/pkg/protobuf", "protobuf")
	f.ImportName("github.com/G-PORTAL/gpcore-cli/pkg/api", "api")
	f.ImportName("google.golang.org/grpc", "grpc")
	f.ImportName("github.com/charmbracelet/ssh", "ssh")
	f.ImportName("github.com/jedib0t/go-pretty/v6/table", "table")
	f.ImportName(apiGRPCImport, apiClient(metadata)+"grpc")

	f.ImportAlias(apiClientImport, apiClient(metadata))
	f.ImportAlias(apiTypesImport, "typesv1")

	// Parameters (variables)
	for _, param := range metadata.Action.Params {
		f.Var().Add(variableDefinition(name, param))
	}
	f.Line()

	// Enum helper functions
	for _, param := range metadata.Action.Params {
		paramType := param.Type
		if isArrayType(param.Type) {
			paramType = strings.TrimPrefix(param.Type, "[]")
		}
		if isEnumType(param.Type) {
			// Add the import for the enum type
			f.ImportAlias(clientPackageName(paramType), stripClient(paramType)+"v"+stripVersion(paramType))

			// Check if we need to add the proto helper
			found := false
			for _, helper := range protoHelpersAdded {
				if helper == param.Type {
					found = true
				}
			}
			if !found {
				alreadyAdded := false
				for _, t := range protoHelpersAdded {
					if t == paramType {
						alreadyAdded = true
					}
				}

				if !alreadyAdded {
					protoHelpersAdded = append(protoHelpersAdded, paramType)
				}
			}
		}
	}

	// Build up the command
	values := Dict{
		Id("Use"):           Lit(metadata.Name),
		Id("Short"):         Lit(metadata.Action.Description),
		Id("Long"):          Lit(metadata.Action.Description),
		Id("SilenceUsage"):  True(),
		Id("SilenceErrors"): True(),
		Id("Args"):          Qual("github.com/spf13/cobra", "OnlyValidArgs"),
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
		if isEnumType(param.Type) && !isArrayType(param.Type) {
			// Enum helper function call
			val = Qual("github.com/G-PORTAL/gpcore-cli/pkg/protobuf", stripPackage(param.Type)+"ToProto").Call(Id(variable))
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

	// Pagination
	if hasListOutput(metadata.Action.APICall) {
		apiCallParams[Id("Pagination")] = Id("pagination")
		c = append(c, Var().Id("totalPages").Int32())
		c = append(c, Id("pagination").Op(":=").Op("&").Qual(apiTypesImport, "PaginationRequest").Values(
			Id("Page").Op(":").Lit(1),
		))
	}

	respC := make([]Code, 0)
	respC = append(respC, List(Id("resp"), Id("err")).Op(":=").
		Id("client").Dot(metadata.Action.APICall.Endpoint).Call(
		Id("cobraCmd").Dot("Context").Call(),
		Op("&").Qual(apiClientImport, metadata.Action.APICall.Endpoint+"Request").Values(apiCallParams)))

	respC = append(respC, If(Id("err").Op("!=").Nil()).Block(
		Return(Id("err"))))

	if hasListOutput(metadata.Action.APICall) {
		// List response
		respC = append(respC, listResponse(metadata)...)

		c = append(c, Var().Id("combinedData").Index().Interface())

		respC = append(respC, Line())
		respC = append(respC, If(Id("resp.Pagination").Op("==").Nil().Block(
			Break())))
		respC = append(respC, Id("totalPages").Op("=").
			Id("resp").Dot("GetPagination").Call().Dot("GetTotal").Call())
		respC = append(respC, Id("pagination").Dot("Page").Op("++"))
		respC = append(respC, If(Id("resp").Dot("Pagination").Dot("Page").Op(">=").Id("totalPages")).Block(
			Break()))
	} else {
		// Single response
		respC = append(respC, singleResponse(metadata)...)
	}

	if hasListOutput(metadata.Action.APICall) {
		// Build the table
		c = append(c, Id("sshSession").Op(":=").
			Id("ctx").Dot("Value").
			Call(Lit("ssh")).
			Assert(Op("*").Qual("github.com/charmbracelet/ssh", "Session")))
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
		c = append(c, Defer().Id("cobraCmd").Dot("SetOut").Call(Nil()))

		if hasListOutput(metadata.Action.APICall) {
			c = append(c, For().Block(respC...))
		} else {
			c = append(c, respC...)
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
	} else {
		c = append(c, respC...)
	}

	respData := "respData"
	if hasListOutput(metadata.Action.APICall) {
		respData = "combinedData"
	}

	jsonOutputFormatCode := make([]Code, 0)
	jsonOutputFormatCode = append(jsonOutputFormatCode,
		List(Id("jsonData"), Id("err")).
			Op(":=").
			Qual("encoding/json", "MarshalIndent").Call(
			Id(respData),
			Lit(""),
			Lit("  ")),
		If(Id("err").Op("!=").Nil()).Block(
			Return(Id("err"))),
		Id("cobraCmd").Dot("Println").Call(
			Id("string").Call(
				Id("jsonData"))))

	if hasListOutput(metadata.Action.APICall) {
		c = append(c,
			If(Qual("github.com/G-PORTAL/gpcore-cli/pkg/config", "JSONOutput")).
				Block(jsonOutputFormatCode...))
	} else {
		c = append(c, jsonOutputFormatCode...)
	}

	c = append(c, Return(Nil()))

	return c
}

// listResponse generates the code for the list response. This function will
// print the response as a table.
func listResponse(metadata SubcommandMetadata) []Code {
	c := make([]Code, 0)

	// Remove root key
	rootKey := metadata.Action.RootKey
	if rootKey == "" {
		rootKey = title(strcase.LowerCamelCase(plural(metadata.Name)))
	}
	c = append(c, Id("respData").Op(":=").
		Id("resp").Dot(rootKey))

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
		allFieldsCode := make([]Code, 0)

		allFieldsCode = append(allFieldsCode, Id("c").Op(":=").
			Qual("reflect", "ValueOf").
			Call(Op("*").Id("entry")))

		if hasListOutput(metadata.Action.APICall) {
			allFieldsCode = append(allFieldsCode, Id("combinedData").Op("=").Append(Id("combinedData"), Id("entry")))
		}
		allFieldsCode = append(allFieldsCode, Id("row").Op(":=").
			Qual("github.com/jedib0t/go-pretty/v6/table", "Row").Values())

		headerCondition := Id("j").Op("==").Lit(0)
		if hasListOutput(metadata.Action.APICall) {
			headerCondition = headerCondition.Op("&&").Id("pagination").Dot("Page").Op("==").Lit(1)
		}

		allFieldsCode = append(allFieldsCode, For(
			Id("i").Op(":=").Lit(0),
			Id("i").Op("<").Id("c").Dot("NumField").Call(),
			Id("i").Op("++")).Block(
			If(Id("c").Dot("Type").Call().Dot("Field").Call(Id("i")).Dot("IsExported").Call().Block(
				Id("field").Op(":=").Id("c").Dot("Field").Call(
					Id("i")).Dot("Interface").Call(),
				Id("val").Op(":=").Lit(""),
				Id("col").Op(":=").Qual("fmt", "Sprintf").Call(
					Lit("%v"),
					Id("c").Dot("Type").Call().
						Dot("Field").Call(
						Id("i")).
						Dot("Name")),
				// TODO: This part should be refactored to a separate function
				Switch(Id("col").Block(
					Case(Lit("CreatedAt")).Block(tableColValue("CreatedAt", "")...),
					Case(Lit("Currency")).Block(tableColValue("Currency", "")...),
					Case(Lit("User")).Block(tableColValue("User", "")...),
					Case(Lit("Environment")).Block(tableColValue("Environment", "")...),
					Case(Lit("Datacenter")).Block(tableColValue("Datacenter", "")...),
					Case(Lit("Company")).Block(tableColValue("Company", "")...),
					Default().Block(
						Id("val").Op("=").Qual("fmt", "Sprintf").Call(
							Lit("%v"), Id("field"),
						),
					))),
				Id("row").Op("=").Append(Id("row"), Id("val")),
				If(headerCondition).Block(
					Id("headerRow").Op("=").Append(
						Id("headerRow"),
						Id("c").Dot("Type").Call().Dot("Field").Call(Id("i")).Dot("Name"))),
			))))
		allFieldsCode = append(allFieldsCode, Id("tbl").Dot("AppendRow").Call(Id("row")))
		allFieldsCode = append(allFieldsCode, If(headerCondition).Block(
			Id("tbl").Dot("AppendHeader").Call(Id("headerRow"))))

		c = append(c, For(List(Id("j"), Id("entry")).
			Op(":=").Range().Id("respData")).Block(allFieldsCode...))
	} else {
		// Header row
		headerCode := make([]Code, 0)

		for _, value := range metadata.Action.Fields {
			colName := value
			if strings.Contains(value, ".") {
				colName = strings.Split(value, ".")[0]
			}
			headerCode = append(headerCode, Id("headerRow").Op("=").Append(Id("headerRow"), Lit(colName)))
		}
		headerCode = append(headerCode, Id("tbl").Dot("AppendHeader").Call(Id("headerRow")))

		if hasListOutput(metadata.Action.APICall) {
			c = append(c, If(Id("pagination").Dot("Page").Op("==").Lit(1)).Block(headerCode...))
		} else {
			c = append(c, headerCode...)
		}
		c = append(c, Line())

		// Only use whitelisted fields
		valuesCode := make([]Code, 0)
		indexVariable := "_"

		valuesCode = append(valuesCode, Id("row").Op(":=").Make(Index().Interface(), Lit(0)))
		valuesCode = append(valuesCode, Id("val").Op(":=").Lit(""))

		// Rows
		for _, value := range metadata.Action.Fields {
			// Format column cell
			valuesCode = append(valuesCode, tableColValue(value, "entry")...)

			// Call hook
			if hasHook(metadata.Definition.Name, metadata.Name, "post") {
				indexVariable = "i"
				valuesCode = append(valuesCode, If(List(Id("v"), Id("ok")).Op(":=").
					Id("respHook").Index(Id("i")).Index(Lit(title(value))).Op(";").Id("ok").Block(
					Id("val").Op("=").Id("v"))))
			}

			// Append to row
			valuesCode = append(valuesCode, Id("row").Op("=").Append(Id("row"), Id("val")))
		}

		// Append to combined data (json output)
		valuesCode = append(valuesCode, Id("tbl").Dot("AppendRow").Call(Id("row")))

		if hasListOutput(metadata.Action.APICall) {
			valuesCode = append(valuesCode, Id("combinedData").Op("=").Append(Id("combinedData"), Id("entry")))
		}
		c = append(c, For(List(Id(indexVariable), Id("entry")).Op(":=").Range().Id("respData")).Block(
			valuesCode...))
	}

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
func tableColValue(col string, structPrefix string) []Code {
	c := make([]Code, 0)

	identifier := col

	// When we have a dot in the identifier, we want to do some special formatting,
	// but we need to extract the identifier first in order to work with it.
	if strings.Contains(col, ".") {
		parts := strings.Split(col, ".")
		identifier = parts[0]
	}

	// If we have a struct prefix, we need to prefix the identifier with it.
	if structPrefix != "" {
		identifier = structPrefix + "." + identifier
	}

	// We define some special "formatting suffixes" for common fields
	// like price, date, time, etc.
	for _, suffix := range api.SpecialFormatters {
		if strings.HasSuffix(col, "."+suffix) {
			col = strings.TrimSuffix(col, "."+suffix)
			c = append(c, Id("val").Op("=").
				Qual("github.com/G-PORTAL/gpcore-cli/pkg/api", "Format"+suffix).Call(
				Id(identifier)))

			return c
		}
	}

	// Special cases
	switch identifier {
	// TODO: Email
	// TODO: FullName
	// TODO: LastLoginAt
	case "CreatedAt":
		c = append(c, Id("val").Op("=").
			Qual("github.com/G-PORTAL/gpcore-cli/pkg/api", "FormatDate").Call(
			Id("field").Assert(Op("*").Qual("google.golang.org/protobuf/types/known/timestamppb", "Timestamp"))))
	case "Currency":
		c = append(c, Id("val").Op("=").Qual("fmt", "Sprintf").Call(
			Lit("%v"), Id("field")).Index(Lit(9), Empty()))
	case "Environment":
		c = append(c, Id("val").Op("=").Qual("fmt", "Sprintf").Call(
			Lit("%v"), Id("field")).Index(Lit(20), Empty()))
	case "User":
		// TODO: Use FormatUser
		c = append(c, Id("user").Op(":=").
			Qual("reflect", "Indirect").Call(
			Id("c")).Dot("FieldByName").Call(Lit("User")))
		c = append(c, Id("val").Op("=").
			Qual("reflect", "Indirect").Call(
			Id("user")).Dot("FieldByName").Call(Lit("Username")).Dot("String").Call())
	case "Company":
		// TODO: Use FormatCompany
		c = append(c, Id("name").Op(":=").
			Qual("reflect", "Indirect").Call(
			Id("c")).Dot("FieldByName").Call(Lit("Company")))
		c = append(c, Id("val").Op("=").
			Qual("reflect", "Indirect").Call(
			Id("name")).Dot("FieldByName").Call(Lit("Name")).Dot("String").Call())
	case "Datacenter":
		// TODO: Use FormatDatacenter
		c = append(c, Id("name").Op(":=").
			Qual("reflect", "Indirect").Call(
			Id("c")).Dot("FieldByName").Call(Lit("Datacenter")))
		c = append(c, Id("val").Op("=").
			Qual("reflect", "Indirect").Call(
			Id("name")).Dot("FieldByName").Call(Lit("Name")).Dot("String").Call())
	default:
		// Default formatting
		c = append(c, Id("val").Op("=").
			Qual("fmt", "Sprintf").Call(
			Lit("%v"),
			Id(identifier)))

	}
	return c
}

// initFunc generates the init function, which will add all flags and params to
// the command and add it to the root command. Required flags are marked as
// such.
func initFunc(name string, metadata SubcommandMetadata) []Code {
	c := make([]Code, 0)

	// Params
	if len(metadata.Action.Params) > 0 {
		for _, param := range metadata.Action.Params {
			dataType := title(param.Type)
			if isEnumType(param.Type) {
				dataType = "String"
			}
			if isArrayType(param.Type) {
				if isEnumType(param.Type) {
					dataType = ""
				} else {
					dataType = title(strings.TrimPrefix(param.Type, "[]")) + "Slice"
				}
			}

			params := make([]Code, 0)
			params = append(params, Op("&").Id(strcase.LowerCamelCase(name)+title(strcase.LowerCamelCase(param.Name))))
			params = append(params, Lit(strcase.KebabCase(param.Name)))

			if !isArrayType(param.Type) || (isArrayType(param.Type) && !isEnumType(param.Type)) {
				params = append(params, defaultValue(param))
			}

			params = append(params, Lit(parameterDescription(param)))

			c = append(c,
				Id(strcase.LowerCamelCase(name)+"Cmd").
					Dot("Flags").Call().
					Dot(dataType+"Var").Params(params...))
		}
		c = append(c, Line())
	}

	// Required fields
	containsRequiredFields := false
	for _, param := range metadata.Action.Params {
		if param.Required {
			c = append(c,
				Id(strcase.LowerCamelCase(name)+"Cmd").
					Dot("MarkFlagRequired").
					Call(Lit(strcase.KebabCase(param.Name))))
			containsRequiredFields = true
		}
	}
	if containsRequiredFields {
		c = append(c, Line())
	}

	// Add the command to the root command when the user has set up the admin
	// configuration.
	addCommand := Id("Root" + strcase.UpperCamelCase(metadata.Definition.Name) + "Command").
		Dot("AddCommand").
		Call(Id(name + "Cmd"))

	if metadata.Action.APICall.Client == "admin" {
		c = append(c, If(
			Qual("github.com/G-PORTAL/gpcore-cli/pkg/config", "HasAdminConfig").Call().Block(
				addCommand)))
	} else {
		c = append(c, addCommand)
	}

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
	case "int32":
		variable.Int32()
	case "int64":
		variable.Int64()
	case "float":
		variable.Float64()
	case "string":
		variable.String()
	// TODO: Fileupload
	default:
		if isArrayType(param.Type) {
			singleType := strings.TrimPrefix(param.Type, "[]") // Remove array indicator

			if isEnumType(param.Type) {
				alreadyAdded := false
				for _, t := range arrayDatatypes {
					if t == singleType {
						alreadyAdded = true
					}
				}
				if !alreadyAdded {
					arrayDatatypes = append(arrayDatatypes, singleType)
				}

				variable.Qual("github.com/G-PORTAL/gpcore-cli/pkg/protobuf", title(stripPackage(singleType))+"Array")
			} else {
				variable.Index().Id(singleType)
			}
		} else {
			if isEnumType(param.Type) {
				variable.String()
			} else {
				variable.Id(title(param.Type))
			}
		}
	}
	return variable
}
