package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/meinside/overwatch-go/stat"
)

const (
	DefaultPlatform = "pc"
	DefaultRegion   = "en"
	DefaultLanguage = "en-us"

	PlatformParamDescription  = `platform string, eg. "pc", ...`
	RegionParamDescription    = `region string, eg. "us", "kr", "eu", ...`
	LanguageParamDescription  = `language string, eg. "en-us", "ko-kr", ...`
	VerboseParamDescription   = `show verbose messages for debugging purpose`
	BattleTagParamDescription = `battle tag, eg. "meinside#3155"`
	ToHtmlParamDescription    = `print html, not json`
	OutFileParamDescription   = `save result to a file`
)

func main() {
	// parse flags
	platform := flag.String("platform", DefaultPlatform, PlatformParamDescription)
	region := flag.String("region", DefaultRegion, RegionParamDescription)
	language := flag.String("language", DefaultLanguage, LanguageParamDescription)
	verbose := flag.Bool("verbose", false, VerboseParamDescription)
	battleTag := flag.String("battletag", "", BattleTagParamDescription)
	toHtml := flag.Bool("html", false, ToHtmlParamDescription)
	outFile := flag.String("out", "", OutFileParamDescription)
	flag.Parse()

	if *battleTag == "" {
		fmt.Printf("* Battle Tag was not given\n")

		flag.PrintDefaults()
	} else {
		stat.Verbose = *verbose

		battleTags := strings.Split(*battleTag, "#")
		if len(battleTags) == 2 {
			if battleTagNumber, err := strconv.Atoi(battleTags[1]); err == nil {
				if result, err := stat.FetchStat(battleTags[0], int(battleTagNumber), *platform, *region, *language); err == nil {
					if *toHtml {
						if html, err := stat.RenderStatToHtml(result, stat.SampleHtmlTemplate); err == nil {
							if *outFile != "" {
								saveToFile(*outFile, []byte(html))
							} else {
								fmt.Printf("%s\n", html)
							}
						} else {
							fmt.Printf("* HTML encode error: %s\n", err)
						}
					} else {
						if bytes, err := json.MarshalIndent(result, "", "\t"); err == nil {
							if *outFile != "" {
								saveToFile(*outFile, bytes)
							} else {
								fmt.Printf("%s\n", string(bytes))
							}
						} else {
							fmt.Printf("* JSON encode error: %s\n", err)
						}
					}
				} else {
					fmt.Printf("* Fetch error: %s\n", err)
				}
			} else {
				fmt.Printf("* Malformed battle tag: %s (%s)\n", *battleTag, err)
			}
		} else {
			fmt.Printf("* Malformed battle tag: %s\n", *battleTag)
		}
	}
}

func saveToFile(filepath string, bytes []byte) error {
	return ioutil.WriteFile(filepath, bytes, 0640)
}
