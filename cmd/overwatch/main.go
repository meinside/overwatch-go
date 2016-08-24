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

	PlatformParamDescription       = `platform string, eg. "pc", ...`
	RegionParamDescription         = `region string, eg. "us", "kr", "eu", ...`
	LanguageParamDescription       = `language string, eg. "en-us", "ko-kr", ...`
	VerboseParamDescription        = `show verbose messages for debugging purpose`
	BattleTagParamDescription      = `battle tag, eg. "meinside#3155"`
	ToHtmlParamDescription         = `print html, not json`
	OutFileParamDescription        = `save result to a file`
	BannerFileParamDescription     = `create a banner file in .png format`
	SuppressOutputParamDescription = `be quiet, no output on stdout`
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
	bannerFile := flag.String("banner", "", BannerFileParamDescription)
	suppressOutput := flag.Bool("quiet", false, SuppressOutputParamDescription)
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
					// print or save result
					if *toHtml {
						if html, err := stat.RenderStatToHtml(result, stat.SampleHtmlTemplate); err == nil {
							if *outFile != "" {
								if err := saveToFile(*outFile, []byte(html)); err != nil {
									fmt.Printf("* Failed to save %s: %s\n", *outFile, err)
								}
							} else {
								if !*suppressOutput {
									fmt.Printf("%s\n", html) // print to stdout
								}
							}
						} else {
							fmt.Printf("* HTML encode error: %s\n", err)
						}
					} else {
						if bytes, err := json.MarshalIndent(result, "", "\t"); err == nil {
							if *outFile != "" {
								if err := saveToFile(*outFile, bytes); err != nil {
									fmt.Printf("* Failed to save %s: %s\n", *outFile, err)
								}
							} else {
								if !*suppressOutput {
									fmt.Printf("%s\n", string(bytes)) // print to stdout
								}
							}
						} else {
							fmt.Printf("* JSON encode error: %s\n", err)
						}
					}

					// if requested, create a banner file
					if *bannerFile != "" {
						if stat.RenderStatToPng(result, *bannerFile); err != nil {
							fmt.Printf("* Failed to create a banner file: %s\n", err)
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
