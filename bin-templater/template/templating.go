package template

import (
	"bufio"
	"encoding/json"
	"github.com/leekchan/gtf"
	"io"
	"os"
	"path"
	"reflect"
	"sort"
	"strings"
	txttmpl "text/template"
	"time"
)

type Templating struct {
	template  *txttmpl.Template
	name      string
	content   string
	functions map[string]interface{}
}

const EXT_CFG = ".cfg"

var TemplateFunctions map[string]interface{}

func NewTemplating(partials *txttmpl.Template, filePath, content string) (*Templating, error) {
	t := Templating{
		name:      filePath,
		content:   CleanupOfTemplate(content),
		functions: TemplateFunctions,
	}
	if partials == nil {
		partials = txttmpl.New(t.name)
	}

	tmpl, err := partials.New(t.name).Funcs(t.functions).Funcs(map[string]interface{}(gtf.GtfFuncMap)).Parse(t.content)
	t.template = tmpl
	return &t, err
}

func CleanupOfTemplate(content string) string {
	var lines []string
	var currentLine string
	scanner := bufio.NewScanner(strings.NewReader(string(content)))
	for scanner.Scan() {
		part := strings.TrimRight(scanner.Text(), " ")
		leftTrim := strings.TrimLeft(part, " ")
		if strings.HasPrefix(leftTrim, "{{-") {
			part = "{{" + leftTrim[3:]
		}
		currentLine += part
		if strings.HasSuffix(currentLine, "-}}") {
			currentLine = currentLine[0:len(currentLine)-3] + "}}"
		} else {
			lines = append(lines, currentLine)
			currentLine = ""
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	return strings.Join(lines, "\n")
}

func (t *Templating) Execute(wr io.Writer, data interface{}) error {
	return t.template.Execute(wr, data)
}

func (t *Templating) AddFunction(name string, fn interface{}) {
	t.functions[name] = fn
}

func (t *Templating) AddFunctions(fs map[string]interface{}) {
	addFuncs(t.functions, fs)
}

///////////////////////////////////

func ifOrDef(eif interface{}, yes interface{}, no interface{}) interface{} {
	if eif != nil {
		return yes
	}
	return no
}

func orDef(val interface{}, def interface{}) interface{} {
	if val != nil {
		return val
	}
	return def
}

func orDefs(val []interface{}, def interface{}) interface{} {
	if val != nil && len(val) != 0 {
		return val
	}
	return []interface{}{def}
}

func addFuncs(out, in map[string]interface{}) {
	for name, fn := range in {
		out[name] = fn
	}
}

func UnmarshalJsonObject(data string) (map[string]interface{}, error) {
	var ret map[string]interface{}
	err := json.Unmarshal([]byte(data), &ret)
	return ret, err
}

func UnmarshalJsonArray(data string) ([]interface{}, error) {
	var ret []interface{}
	err := json.Unmarshal([]byte(data), &ret)
	return ret, err
}

func IsType(data interface{}, t string) bool {
	dataType := reflect.TypeOf(data)
	if dataType == nil {
		return false
	}
	if dataType.String() == t {
		return true
	}
	return false
}

func IsKind(data interface{}, t string) bool {
	dataType := reflect.TypeOf(data)
	if dataType == nil {
		return false
	}
	if dataType.Kind().String() == t {
		return true
	}
	return false
}

func IsMap(data interface{}) bool {
	dataType := reflect.TypeOf(data)
	if dataType == nil {
		return false
	}
	if dataType.Kind() == reflect.Map {
		return true
	}
	return false
}

func IsArray(data interface{}) bool {
	dataType := reflect.TypeOf(data)
	if dataType == nil {
		return false
	}
	if dataType.Kind() == reflect.Array || dataType.Kind() == reflect.Slice {
		return true
	}
	return false
}

func IsString(data interface{}) bool {
	dataType := reflect.TypeOf(data)
	if dataType == nil {
		return false
	}
	if dataType.Kind() == reflect.String {
		return true
	}
	return false
}

func IsMapFirst(data interface{}, element interface{}) bool {
	switch reflect.TypeOf(data).Kind() {
	case reflect.Map:
		mapItem := reflect.ValueOf(data).MapKeys()

		var keys []string
		for _, k := range mapItem {
			keys = append(keys, k.String())
		}
		sort.Strings(keys)
		mapItemType := keys[0]
		return (mapItemType == element)
	}
	return false
}

func IsMapLast(data interface{}, element interface{}) bool {
	switch reflect.TypeOf(data).Kind() {
	case reflect.Map:
		mapItem := reflect.ValueOf(data).MapKeys()

		var keys []string
		for _, k := range mapItem {
			keys = append(keys, k.String())
		}
		sort.Strings(keys)
		mapItemType := keys[len(keys)-1]
		return (mapItemType == element)
	}
	return false
}

func HowDeep(data interface{}, element interface{}) int {
	return HowDeepIsIt(data, element, 0)
}

func HowDeepIsIt(data interface{}, element interface{}, deep int) int {
	elemType := reflect.TypeOf(element).Kind()
	mapItem := reflect.ValueOf(data)
	elemItem := reflect.ValueOf(element)
	switch elemType {
	case reflect.Map:
		for _, b := range reflect.ValueOf(data).MapKeys() {
			if reflect.DeepEqual(mapItem.MapIndex(b).Interface(), elemItem.Interface()) {
				return deep + 1
			}
			if IsMap(mapItem.MapIndex(b).Interface()) {
				index := HowDeepIsIt(mapItem.MapIndex(b).Interface(), element, deep+1)
				if index == deep+2 {
					return index
				}
			}
		}
	}

	return deep
}

func add(x, y int) int {
	return x + y
}

func mul(x, y int) int {
	return x * y
}

func div(x, y int) int {
	return x / y
}

func mod(x, y int) int {
	return x % y
}

func sub(x, y int) int {
	return x - y
}

func init() {
	TemplateFunctions = make(map[string]interface{})
	TemplateFunctions["base"] = path.Base
	TemplateFunctions["split"] = strings.Split
	TemplateFunctions["json"] = UnmarshalJsonObject
	TemplateFunctions["jsonArray"] = UnmarshalJsonArray
	TemplateFunctions["dir"] = path.Dir
	TemplateFunctions["getenv"] = os.Getenv
	TemplateFunctions["join"] = strings.Join
	TemplateFunctions["datetime"] = time.Now
	TemplateFunctions["toUpper"] = strings.ToUpper
	TemplateFunctions["toLower"] = strings.ToLower
	TemplateFunctions["contains"] = strings.Contains
	TemplateFunctions["replace"] = strings.Replace
	TemplateFunctions["repeat"] = strings.Repeat
	TemplateFunctions["orDef"] = orDef
	TemplateFunctions["orDefs"] = orDefs
	TemplateFunctions["ifOrDef"] = ifOrDef
	TemplateFunctions["isType"] = IsType
	TemplateFunctions["isMap"] = IsMap
	TemplateFunctions["isArray"] = IsArray
	TemplateFunctions["isKind"] = IsKind
	TemplateFunctions["isString"] = IsString
	TemplateFunctions["isMapFirst"] = IsMapFirst
	TemplateFunctions["isMapLast"] = IsMapLast
	TemplateFunctions["howDeep"] = HowDeep
	TemplateFunctions["add"] = add
	TemplateFunctions["mul"] = mul
	TemplateFunctions["div"] = div
	TemplateFunctions["sub"] = sub
	TemplateFunctions["mod"] = mod

	TemplateFunctions["IsMapFirst"] = IsMapFirst
	TemplateFunctions["IsMapLast"] = IsMapLast
	TemplateFunctions["HowDeep"] = HowDeep
}
