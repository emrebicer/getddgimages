# getddgimages
Download images that are listed on the duckduckgo search engine.

## Install
```terminal
go get github.com/emrebicer/getddgimages
```

## Example usage
This code snippet will download `5` `golang` related images and save it under the `{cwd}/golang` folder.
```go
package main

import (
        "fmt"
        gdi "github.com/emrebicer/getddgimages"
)

func main() {
        result, downloadErr := gdi.DownloadImages("golang", 5)
        if downloadErr != nil {
                fmt.Printf("Failed -> %s\n", downloadErr)
                return
        }
        
        fmt.Printf("Downloaded %d files, the paths are;\n", len(result))
        for k, v := range result {
                fmt.Printf("%d -> %s\n", k+1, v)
        }
}
```
