package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/alecthomas/participle"
	"github.com/iancoleman/strcase"
)

type Object struct {
	Includes []*Include `@@*`
	Types    []*Type    `@@ { @@ }`
}

type Include struct {
	Value string `"include" @String`
}

type Type struct {
	Name       string      `@Ident "{"`
	Functions  []*Function `"function" @@*`
	Signals    []*Signal   `"signal" @@*`
	Properties []*Property `@@*`
}

type Function struct {
	Signature string `@String ":"`
	Return    string `@String`
}

type Signal struct {
	Signature string `@String`
}

type Property struct {
	Type string `@("bool"|"qint8"|"qint16"|"qint32"|"qint64"|"quint8"|"quint16"|"quint32"|"quint64"|"float"|"double"|"QBitArray"|"QBrush"|"QByteArray"|"QColor"|"QCursor"|"QDate"|"QDateTime"|"QEasingCurve"|"QFont"|"QGenericMatrix"|"QIcon"|"QImage"|"QKeySequence"|"QMargins"|"QMatrix4x4"|"QPalette"|"QPen"|"QPicture"|"QPixmap"|"QPoint"|"QQuaternion"|"QRect"|"QRegExp"|"QRegularExpression"|"QRegion"|"QSize"|"QString"|"QTime"|"QTransform"|"QUrl"|"QVariant"|"QVector2D"|"QVector3D"|"QVector4D")* `
	Name string `@Ident "}"`
}

func prettyPrint(i interface{}) {
	s, _ := json.MarshalIndent(i, "", "\t")
	println(string(s))
}

func main() {
	if len(os.Args) < 2 {
		println("Please provide an input file")
		return
	}
	bytes, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	parser, _ := participle.Build(&Object{})

	ast := Object{}
	err = parser.ParseBytes(bytes, &ast)
	if err != nil {
		println(err.Error())
		os.Exit(3)
	}
	fmt.Printf("#pragma once\n")
	for _, val := range ast.Includes {
		fmt.Printf(`#include "%s"`+"\n", val.Value)
	}
	fmt.Printf("\n")
	for _, types := range ast.Types {
		fmt.Printf("class %s : public QObject {\n\tQ_OBJECT\npublic:\n", types.Name)
		fmt.Printf("\texplicit %s(QObject *parent = nullptr);\n", types.Name)

		for _, property := range types.Properties {
			fmt.Printf("\tQ_PROPERTY(%s %s MEMBER m_%s NOTIFY %sChanged)\n", property.Type, property.Name, strcase.ToLowerCamel(property.Name), strcase.ToLowerCamel(property.Name))
		}

		fmt.Printf("\n")
		for _, function := range types.Functions {
			fmt.Printf("\tQ_INVOKABLE %s %s;\n", function.Return, function.Signature)
		}

		fmt.Printf("\nsignals:\n")
		for _, property := range types.Properties {
			fmt.Printf("\tvoid %sChanged(%s val);\n", strcase.ToLowerCamel(property.Name), property.Type)
		}

		fmt.Printf("\n")
		for _, signal := range types.Signals {
			fmt.Printf("\tvoid %s;\n", signal.Signature)
		}

		fmt.Printf("\nprivate:\n")
		for _, property := range types.Properties {
			fmt.Printf("\t%s m_%s;\n", property.Type, strcase.ToLowerCamel(property.Name))
		}
		fmt.Printf("};\n")
	}
}
