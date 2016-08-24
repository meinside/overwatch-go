# overwatch-go

Codes for fetching play stats of [Overwatch™](https://playoverwatch.com).

Stats are crawled from [Blizzard](https://www.blizzard.com)'s official site, and parsed with [Gokogiri](https://github.com/jbowtie/gokogiri).

The result is rendered into JSON format, so it can be used easily in other applications or services.

When it suddenly stops working, there may be some changes on the web site, so please let me know.

## install

```bash
$ go get -u github.com/meinside/overwatch-stats/...
```

## run command

```bash
# default platform: "pc", region: "us", language: "en-us"
$ overwatch -battletag "meinside#3155"
#
$ overwatch -platform pc -region kr -language "ko-kr" -battletag "meinside#3155"
# save to a html file
$ overwatch -region kr -language ko-kr -battletag "meinside#3155" -html -out "/tmp/test_output.html"
```

## sample usage

```go
// it will fetch my stats json and print it to screen
// (https://playoverwatch.com/ko-kr/career/pc/kr/meinside-3155)
package main

import (
	"encoding/json"
	"fmt"

	"github.com/meinside/overwatch-go/stat"
)

func main() {
	if stat, err := stat.FetchStat("meinside", 3155, "pc", "kr", "ko-kr"); err == nil {
		if bytes, err := json.MarshalIndent(stat, "", "\t"); err == nil {
			// print json response
			fmt.Printf("%s\n", string(bytes))
		} else {
			fmt.Printf("* JSON encode error: %s\n", err)
		}
	} else {
		fmt.Printf("* Fetch error: %s\n", err)
	}
}
```

## license

MIT
