# Translate
Translate library for golang!

![Test](https://github.com/vcapi/translate/workflows/Test/badge.svg)

### Usage
Translate from english to chinese
```go
package main

import (
    "context"
    "fmt"

    "github.com/vcapi/translate"
)

func main() {
    input := "How old are you?"
    sLang := "en"
    tLang := "zh-CN"
    ctx := context.TODO()
    val, err := translate.Google(ctx, input, sLang, tLang)
    if err != nil {
        panic(err)
    }
    fmt.Println(val)
}
```


### LICENSE
MIT License

