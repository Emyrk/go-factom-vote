package common

import (
	"fmt"
	"reflect"
	"regexp"
	"time"

	"strings"

	"crypto/sha512"

	"crypto/hmac"
	"crypto/sha256"

	"hash"

	"crypto/md5"
	"crypto/sha1"

	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
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

func computeSha512(data []byte) []byte {
	h := sha512.New()
	h.Write(data)
	return h.Sum(nil)
}

// CheckMAC reports whether messageMAC is a valid HMAC tag for message.
func CheckMAC(algo string, message, messageMAC, key []byte) bool {
	var f func() hash.Hash
	switch algo {
	case "sha256":
		f = sha256.New
	case "sha512":
		f = sha512.New
	case "sha1":
		f = sha1.New
	case "md5":
		f = md5.New
	default:
		log.WithFields(log.Fields{"pkg": "utls", "algo": algo}).Errorf("Hmac algo not supported")
		return false
	}

	mac := hmac.New(f, key)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}
