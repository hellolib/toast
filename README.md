# Toast

[![license](https://img.shields.io/github/license/hellolib/toast)](https://github.com/hellolib/toast/blob/master/LICENSE)

Cross-platform library for sending desktop notifications

## Installation

```bash
go get github.com/hellolib/toast
```

## Example

- Common invocation
  ```go
  package main
  
  import (
      "github.com/hellolib/toast"
  )
  
  func main() {
      // _ = toast.Push("test message")
      _ = toast.Push("test message", toast.WithTitle("app title"))
  }
  
  ```

- `macOS`
    ```go
    package main
    
    import (
        "github.com/hellolib/toast"
    )
    
    func main() {
        // _ = toast.Push("test message")
        // _ = toast.Push("test message", toast.WithTitle("app title"))
        _ = toast.Push("test message",
            toast.WithTitle("app title"),
            toast.WithSubtitle("app sub title"),
            toast.WithAudio(toast.Ping),
            // toast.WithObjectiveC(true),
        )
    }
    
    ```

- `Windows`
  ```go
  package main
  
  import (
      "github.com/hellolib/toast"
  )
  
  func main() {
      // _ = toast.Push("test message")
      // _ = toast.Push("test message", toast.WithTitle("app title"))
      _ = toast.Push("test message",
          toast.WithTitle("app title"),
          toast.WithAppID("app id"),
          toast.WithAudio(toast.Default),
          toast.WithLongDuration(),
          toast.WithIcon("/path/icon.png"),
      )
      // bs, err := os.ReadFile("/path/icon.png")
      // if err != nil {
      // 	log.Fatalln(err)
      // }
      // toast.WithIconRaw(bs)
  }
  
  ```

- `js && wasm`
  ```go
  package main
  
  import (
      "fmt"
      "github.com/hellolib/toast"
  )
  
  func main() {
      // _ = toast.Push("test message")
      // _ = toast.Push("test message", toast.WithTitle("app title"))
      _ = toast.Push("test_message",
          toast.WithTitle("GO-WASM-APP"),
          toast.WithOnClick(func(event interface{}) {
              fmt.Println("click")
          }),
          toast.WithOnClose(func() {
              fmt.Println("close")
          }),
          toast.WithOnShow(func() {
              fmt.Println("show")
          }),
      )
  }
    
  ```
