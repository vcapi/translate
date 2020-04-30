# Translate
Translate library for golang!



### Usage
Translate from english to chinese
```go
package main

import (
    "fmt"

    "github.com/vcapi/translate"
)

func main() {
    input := "How old are you?"
    sLang := "en"
    tLang := "zh-CN"
    val, err := translate.Google(input, sLang, tLang)
    if err != nil {
        panic(err)
    }
    fmt.Println(val)
}
```


### LICENSE
MIT License

