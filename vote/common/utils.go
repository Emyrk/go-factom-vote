package common

import (
	"fmt"
	"reflect"
	"regexp"
	"time"

	"strings"

	"github.com/lib/pq"
)

var paramstring = regexp.MustCompile(`,\s*`)

type ISQLObject interface {
	SelectRows() string
	RowValuePointers() []interface{}

	New() ISQLObject
	Table() string
	InsertFunction() string
}

func InsertQueryParams(block ISQLObject) string {
	return fmt.Sprintf(SelectParamRows(block.SelectRows()), RowValuesFromPointer(block.RowValuePointers())...)
}

func SelectParamRows(selectRow string) string {
	addparam := "param_" + string(paramstring.ReplaceAll([]byte(selectRow), []byte(",\n param_")))
	return strings.Replace(addparam, ",", " := %v,", -1) + ":= %v "
}

func RowValuesFromPointer(pointers []interface{}) []interface{} {
	vals := make([]interface{}, len(pointers))
	for i, p := range pointers {
		v := reflect.ValueOf(p).Elem()
		ptype := reflect.TypeOf(p).String()
		if ptype == "*string" {
			vals[i] = fmt.Sprintf("'%v'", v)
			continue
		} else if ptype == "*time.Time" {
			t := p.(*time.Time)
			vals[i] = fmt.Sprintf("'%s'", string(pq.FormatTimestamp(*t)))
			continue
		} else if ptype == "*[][]uint8" {
			a := p.(*[][]byte)
			sep := ""
			str := "array["
			for _, b := range *a {
				str += sep
				str += fmt.Sprintf("decode('%x', 'hex')", b[:])
				sep = ","
			}
			str += "] :: bytea[]"
			vals[i] = str
			continue
		} else if ptype == "*[]uint8" {
			b := *(p.(*[]byte))
			vals[i] = fmt.Sprintf("decode('%x', 'hex')", b[:])
			continue
		}
		vals[i] = v
	}
	return vals
}

type SQLRowWithScan interface {
	Scan(dest ...interface{}) error
}
