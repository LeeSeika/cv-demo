## JSONX

一个能够动态解析 JSON 数据的 Go 语言包。

#### 背景
因为项目中存在大量动态 JSON 结构的场景，使用传统的静态结构体定义和解析方式会导致代码臃肿且难以维护。

#### Rust 原型
JSONX 的设计灵感来源于 Rust 的 Serde 库，Serde 提供了一种名为 serde_json::Value 的动态数据结构，可以灵活地表示任意 JSON 数据。

``` rust
/// Represents any valid JSON value.
///
/// See the [`serde_json::value` module documentation](self) for usage examples.
#[derive(Clone, Eq, PartialEq)]
pub enum Value {
    /// Represents a JSON null value.
    Null,

    /// Represents a JSON boolean.
    Bool(bool),

    /// Represents a JSON number, whether integer or floating point.
    Number(Number),

    /// Represents a JSON string.
    String(String),

    /// Represents a JSON array.
    Array(Vec<Value>),

    /// Represents a JSON object.
    Object(Map<String, Value>),
}
```

serde_json::Value 提供了丰富的方法用于访问和操作 JSON 数据，例如通过索引访问数组元素和对象属性、类型判断等。

``` rust
impl Value {
    // 根据索引获取对应的对象属性或数组元素
    /// #
    /// let object = json!({ "A": 65, "B": 66, "C": 67 });
    /// assert_eq!(*object.get("A").unwrap(), json!(65));
    ///
    pub fn get<I: Index>(&self, index: I) -> Option<&Value> {
        index.index_into(self)
    }
    // 判断是否为字符串类型
    pub fn is_string(&self) -> bool {
        self.as_str().is_some()
    }
    // 转换为字符串引用
    pub fn as_str(&self) -> Option<&str> {
        match self {
            Value::String(s) => Some(s),
            _ => None,
        }
    }
    // 判断是否为数字类型
    pub fn is_number(&self) -> bool {
        match *self {
            Value::Number(_) => true,
            _ => false,
        }
    }
    // 转换为数字引用
    pub fn as_number(&self) -> Option<&Number> {
        match self {
            Value::Number(number) => Some(number),
            _ => None,
        }
    }
}
```

我们可以将结构体中的动态 JSON 字段类型定义为 serde_json::Value，从而实现对动态 JSON 数据的灵活处理。

``` rust
use serde::{Deserialize, Serialize};
use serde_json::Value as JSONValue;

#[derive(Serialize, Deserialize, PartialEq, Debug)]
pub(crate) struct CatalogList {
    id: String,
    label: JSONValue,
}

impl WaferSettings for CatalogList {
    #[cfg(feature = "schema-locale")]
    fn set_locale(&mut self, provider: &dyn LocaleSchemaProvider, locale: &str) {
        // handle label
        if self.label.is_string() {
            let label_str = self.label.as_str().unwrap_or("");
            // xxxxxx
        } else if self.label.is_object() {
            if let Some(label_value) = self.label.get(locale) {
                // xxxxxx
            } 
        } else {
            // xxxxxx
        }
    }
}
```

#### Go 实现

在 Go 中，存在类似的 JSON 工具包，例如 [tidwall/gjson](https://github.com/tidwall/gjson)，它们也提供了动态操作 JSON 数据的能力。

``` go
// Result represents a json value that is returned from Get().
type Result struct {
	// Type is the json type
	Type Type
	// Raw is the raw json
	Raw string
	// Str is the json string
	Str string
	// Num is the json number
	Num float64
	// Index of raw value in original json, zero means index unknown
	Index int
	// Indexes of all the elements that match on a path containing the '#'
	// query character.
	Indexes []int
}

// IsBool returns true if the result value is a JSON boolean.
func (t Result) IsBool() bool {
	return t.Type == True || t.Type == False
}
// Bool returns an boolean representation.
func (t Result) Bool() bool {
	switch t.Type {
	default:
		return false
	case True:
		return true
	case String:
		b, _ := strconv.ParseBool(strings.ToLower(t.Str))
		return b
	case Number:
		return t.Num != 0
	}
}
```


在 Rust 中，Serde 对 serde_json::Value 整个枚举类型实现了 Serialize 和 Deserialize trait，所以可以直接用于结构体字段的序列化和反序列化，同时也能通过各种方法动态地访问和操作 JSON 数据。<br>

由于语言特性和开源实现上的差异，我们并不能将 gjson 中的 Result 结构体直接用作动态 JSON 的字段类型。<br>

在 Go 里面，如果我们想要对某个字段实现“延迟解析”，通常的做法是将该字段定义为 json.RawMessage 类型，然后在需要使用该字段时再进行解析。<br>

``` go
type Fruit struct {
    Type    string          `json:"type"`
    Data    json.RawMessage `json:"data"`
}
```

现在，我们既想要实现“延迟解析”，又想要方便地访问和操作动态 JSON 数据，我们运用`组合`的方式，定义了一个新的 JSONValue 类型，集成了 json.RawMessage 和 gjson.Result 的功能。<br>

``` go
type JSONValue struct {
    json.RawMessage
    res *gjson.Result    `json:"-"`
}

func (jv *JSONValue) Result() gjson.Result {
	if jv.res != nil {
		return *jv.res
	}
	res := gjson.ParseBytes(jv.RawMessage)
	jv.res = &res
	return res
}
```
*gjson.Result 通过指针引用的方式集成在 JSONValue 结构体中，并且定义为私有字段，通过 Result() 方法懒加载，避免了不必要的内存拷贝开销。<br>

字段的真实数据存储在 json.RawMessage 中，当我们需要动态地访问和操作 JSON 数据时，可以通过 gjson 解析 RawMessage，并将结果缓存到 res 字段中。<br>

res 字段打上了 `json:"-"` 标签，不参与序列化和反序列化过程。<br>

通过这种方式，我们在 Go 中实现了类似于 Rust Serde 中 serde_json::Value 的动态 JSON 处理能力，同时也保留了“延迟解析”的特性。

``` go
var inputJSON []byte = []byte(`
{
    "fruits": [
        {
            "type": "apple",
            "color": "red",
            "peeled": false
        },
        {
            "type": "watermelon",
            "sliced": true,
            "has_seeds": false
        }
    ]
}
`)

type Basket struct {
	Fruits []jsonx.JSONValue `json:"fruits"`
}

type Apple struct {
	Type   string `json:"type"`
	Color  string `json:"color"`
	Peeled bool   `json:"peeled"`
}

type Watermelon struct {
	Type     string `json:"type"`
	Sliced   bool   `json:"sliced"`
	HasSeeds bool   `json:"has_seeds"`
}

func TestLoadFruitsFromJSON(t *testing.T) {
	var basket Basket
	if err := json.Unmarshal(inputJSON, &basket); err != nil {
		t.Fatalf("failed to unmarshal input JSON: %v", err)
	}

	for _, fruit := range basket.Fruits {
		switch typ := fruit.Result().Get("type").String(); typ {
		case "apple":
			var apple Apple
			if err := json.Unmarshal(fruit.RawMessage, &apple); err != nil {
				t.Errorf("failed to unmarshal apple: %v", err)
				continue
			}
			t.Logf("Loaded Apple: %+v", apple)
		case "watermelon":
			var watermelon Watermelon
			if err := json.Unmarshal(fruit.RawMessage, &watermelon); err != nil {
				t.Errorf("failed to unmarshal watermelon: %v", err)
				continue
			}
			t.Logf("Loaded Watermelon: %+v", watermelon)
		default:
			t.Errorf("unknown fruit type: %s", typ)
		}
	}
}
```

