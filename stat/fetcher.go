package stat

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"

	// $ sudo apt-get install libxml2-dev libonig-dev
	//
	// XXX - https://github.com/moovweb/gokogiri/issues/92
	// $ go get -u github.com/jbowtie/gokogiri
	"github.com/jbowtie/gokogiri"
	"github.com/jbowtie/gokogiri/css"
	"github.com/jbowtie/gokogiri/html"
	"github.com/jbowtie/gokogiri/xml"
)

const (
	FakeUserAgent = "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_6) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/52.0.2743.82 Safari/537.36"
)

const (
	NoCompetitiveRank = -1
)

var Verbose bool = false

// get bytes using HTTP GET
func httpGet(url string) (bytes []byte, err error) {
	client := &http.Client{}

	var req *http.Request
	var res *http.Response

	if req, err = http.NewRequest("GET", url, nil); err == nil {
		req.Header.Set("User-Agent", FakeUserAgent)

		if res, err = client.Do(req); err == nil {
			defer res.Body.Close()

			if res.StatusCode == 200 {
				bytes, err = ioutil.ReadAll(res.Body)
			} else {
				err = fmt.Errorf("HTTP %d on GET request for: %s", res.StatusCode, url)
			}
		}
	}

	return bytes, err
}

// fetch given user's stat from official overwatch site.
//
// ex: https://playoverwatch.com/en-us/career/pc/kr/meinside-3155
func FetchStat(battleTagString string, battleTagNumber int, platform, region, language string) (result Stat, err error) {
	url := fmt.Sprintf("https://playoverwatch.com/%s/career/%s/%s/%s-%d",
		language,
		platform,
		region,
		battleTagString,
		battleTagNumber,
	)

	if Verbose {
		log.Printf("> fetching from url: %s\n", url)
	}

	// fetch html content,
	var bytes []byte
	bytes, err = httpGet(url)
	if err != nil {
		return Stat{}, err
	}

	if Verbose {
		log.Printf("> received html: %s\n", string(bytes))
	}

	// parse it and assign to struct
	return parseStat(bytes)
}

// parse stat from html bytes
//
// XXX - if it stops working, should check the html response and alter css selectors
func parseStat(bytes []byte) (result Stat, err error) {
	var doc *html.HtmlDocument
	if doc, err = gokogiri.ParseHtml(bytes); err == nil {
		defer doc.Free()

		////////////////
		// [info]
		//
		var name string
		if name, err = extractString(doc, "div.masthead-player > h1.header-masthead"); err != nil {
			return Stat{}, err
		}
		var profileImageUrl string
		if profileImageUrl, err = extractFirstAttrString(doc, "div.masthead-player > img.player-portrait", "src"); err != nil {
			return Stat{}, err
		}
		var level int32
		if level, err = extractInt32(doc, "div.player-level > div"); err != nil {
			return Stat{}, err
		}
		var competitiveRank int32
		if competitiveRank, err = extractInt32(doc, "div.competitive-rank > div"); err != nil {
			competitiveRank = NoCompetitiveRank
		}
		//
		////////////////
		// [stats] quick play
		//
		var featuredStats map[string]string
		var topHeroes map[string][]Hero
		var careerStats []CareerStat
		if featuredStats, topHeroes, careerStats, err = extractPlayStat(doc, TagIdQuickPlay); err != nil {
			return Stat{}, err
		}
		quickPlayStat := PlayStat{
			FeaturedStats: featuredStats,
			TopHeroes:     topHeroes,
			CareerStats:   careerStats,
		}
		//
		// [stats] competitive play
		//
		var competitivePlayStat PlayStat
		if competitiveRank != NoCompetitiveRank {
			if featuredStats, topHeroes, careerStats, err = extractPlayStat(doc, TagIdCompetitivePlay); err != nil {
				return Stat{}, err
			}
			competitivePlayStat = PlayStat{
				FeaturedStats: featuredStats,
				TopHeroes:     topHeroes,
				CareerStats:   careerStats,
			}
		} else {
			competitivePlayStat = PlayStat{}
		}
		//
		////////////////
		// achievements
		//
		achievements := []AchievementCategory{}
		var achievementCategoryNames []string
		if achievementCategoryNames, err = extractStrings(doc, "#achievements-section select > option"); err != nil {
			return Stat{}, err
		}
		for i, categoryName := range achievementCategoryNames {
			var urls, titles, descriptions, classes []string

			// achieved/non-achieved achievements
			achieved := []Achievement{}
			nonAchieved := []Achievement{}
			if urls, err = extractAttrStrings(doc, fmt.Sprintf("#achievements-section div[data-group-id=\"achievements\"]:nth-of-type(%d) > ul div.achievement-card > img", i+1), "src"); err != nil {
				return Stat{}, err
			}
			if titles, err = extractStrings(doc, fmt.Sprintf("#achievements-section div[data-group-id=\"achievements\"]:nth-of-type(%d) div.tooltip-tip > h6", i+1)); err != nil {
				return Stat{}, err
			}
			if descriptions, err = extractStrings(doc, fmt.Sprintf("#achievements-section div[data-group-id=\"achievements\"]:nth-of-type(%d) div.tooltip-tip > p", i+1)); err != nil {
				return Stat{}, err
			}
			if classes, err = extractAttrStrings(doc, fmt.Sprintf("#achievements-section div[data-group-id=\"achievements\"]:nth-of-type(%d) > ul div.achievement-card", i+1), "class"); err != nil {
				return Stat{}, err
			}
			for i, class := range classes {
				if strings.Contains(class, "m-disabled") { // m-disabled: non-achieved achievement
					nonAchieved = append(nonAchieved, Achievement{
						Title:       titles[i],
						Description: descriptions[i],
						ImageUrl:    urls[i],
					})
				} else {
					achieved = append(achieved, Achievement{
						Title:       titles[i],
						Description: descriptions[i],
						ImageUrl:    urls[i],
					})
				}
			}

			achievements = append(achievements, AchievementCategory{
				Name:        categoryName,
				Achieved:    achieved,
				NonAchieved: nonAchieved,
			})
		}

		// return result
		return Stat{
			Name:            name,
			ProfileImageUrl: profileImageUrl,
			Level:           level,
			CompetitiveRank: competitiveRank,

			QuickPlay:       quickPlayStat,
			CompetitivePlay: competitivePlayStat,

			Achievements: achievements,
		}, err
	}

	return Stat{}, err
}

func extractPlayStat(doc *html.HtmlDocument, id TagId) (featuredStats map[string]string, topHeroes map[string][]Hero, careerStats []CareerStat, err error) {
	featuredStats = make(map[string]string)
	topHeroes = make(map[string][]Hero)
	careerStats = []CareerStat{}

	////////////////
	// featured stats
	var featuredStatTitles []string
	if featuredStatTitles, err = extractStrings(doc, fmt.Sprintf("#%s > section.highlights-section div.card-content > p", id)); err != nil {
		return featuredStats, topHeroes, careerStats, err
	}
	for i, title := range featuredStatTitles {
		var value string
		if value, err = extractString(doc, fmt.Sprintf("#%s > section.highlights-section li:nth-child(%d) div.card-content > h3", id, i+1)); err != nil {
			return featuredStats, topHeroes, careerStats, err
		}
		featuredStats[title] = value
	}
	//
	////////////////
	// top heroes
	var comparison, heroNames, heroImageUrls, heroValues []string
	if comparison, err = extractStrings(doc, fmt.Sprintf("#%s > section.hero-comparison-section select[data-group-id=\"comparisons\"] > option", id)); err != nil {
		return featuredStats, topHeroes, careerStats, err
	}
	for i, comparison := range comparison {
		heroes := []Hero{}
		if heroNames, err = extractStrings(doc, fmt.Sprintf("#%s > section.hero-comparison-section div.progress-category:nth-of-type(%d) div.bar-text > div.title", id, i+1)); err != nil {
			return featuredStats, topHeroes, careerStats, err
		}
		if heroImageUrls, err = extractAttrStrings(doc, fmt.Sprintf("#%s > section.hero-comparison-section div.progress-category:nth-of-type(%d) img", id, i+1), "src"); err != nil {
			return featuredStats, topHeroes, careerStats, err
		}
		if heroValues, err = extractStrings(doc, fmt.Sprintf("#%s > section.hero-comparison-section div.progress-category:nth-of-type(%d) div.bar-text > div.description", id, i+1)); err != nil {
			return featuredStats, topHeroes, careerStats, err
		}
		for i, _ := range heroNames {
			heroes = append(heroes, Hero{
				Name:     heroNames[i],
				ImageUrl: heroImageUrls[i],
				Value:    heroValues[i],
			})
		}

		topHeroes[comparison] = heroes
	}
	//
	////////////////
	// career stats
	//
	var statIds []string
	if statIds, err = extractAttrStrings(doc, fmt.Sprintf("#%s div[data-group-id=\"stats\"]", id), "data-category-id"); err != nil {
		return featuredStats, topHeroes, careerStats, err
	}
	for _, statId := range statIds {
		var heroName string
		if heroName, err = extractString(doc, fmt.Sprintf("#%s option[value=\"%s\"]", id, statId)); err != nil {
			return featuredStats, topHeroes, careerStats, err
		}
		var categoryNames []string
		if categoryNames, err = extractStrings(doc, fmt.Sprintf("#%s div[data-category-id=\"%s\"] div.card-stat-block > table.data-table > thead > tr > th > span.stat-title", id, statId)); err != nil {
			return featuredStats, topHeroes, careerStats, err
		}

		var categoryAttrs, categoryValues []string
		careerStatCategories := []CareerStatCategory{}
		for i, categoryName := range categoryNames {
			if categoryAttrs, err = extractStrings(doc, fmt.Sprintf("#%s div[data-category-id=\"%s\"] > div:nth-child(%d) > div.card-stat-block > table.data-table > tbody > tr > td:nth-child(1)", id, statId, i+1)); err != nil {
				return featuredStats, topHeroes, careerStats, err
			}
			if categoryValues, err = extractStrings(doc, fmt.Sprintf("#%s div[data-category-id=\"%s\"] > div:nth-child(%d) > div.card-stat-block > table.data-table > tbody > tr > td:nth-child(2)", id, statId, i+1)); err != nil {
				return featuredStats, topHeroes, careerStats, err
			}

			// values for this category
			values := map[string]string{}
			for i, _ := range categoryAttrs {
				values[categoryAttrs[i]] = categoryValues[i]
			}

			// categories
			careerStatCategories = append(careerStatCategories, CareerStatCategory{
				Name:   categoryName,
				Values: values,
			})
		}

		// career stats for each hero
		careerStats = append(careerStats, CareerStat{
			HeroName:   heroName,
			Categories: careerStatCategories,
		})
	}

	return featuredStats, topHeroes, careerStats, nil
}

func extractInt32(doc *html.HtmlDocument, selector string) (int32, error) {
	if s, err := extractString(doc, selector); err == nil {
		if i, err := strconv.ParseInt(strings.Replace(s, ",", "", -1), 10, 32); err == nil { // XXX - remove unwanted ','
			return int32(i), nil
		} else {
			return 0, err
		}
	} else {
		return 0, err
	}
}

func extractInt64(doc *html.HtmlDocument, selector string) (int64, error) {
	if s, err := extractString(doc, selector); err == nil {
		return strconv.ParseInt(strings.Replace(s, ",", "", -1), 10, 64) // XXX - remove unwanted ','
	} else {
		return 0, err
	}
}

func extractFloat32(doc *html.HtmlDocument, selector string) (float32, error) {
	if s, err := extractString(doc, selector); err == nil {
		if f, err := strconv.ParseFloat(s, 32); err == nil {
			return float32(f), nil
		} else {
			return 0, err
		}
	} else {
		return 0, err
	}
}

func extractString(doc *html.HtmlDocument, selector string) (string, error) {
	xpath := css.Convert(selector, css.GLOBAL)

	if searched, err := doc.Search(xpath); err == nil {
		if len(searched) > 0 {
			return searched[0].Content(), nil // take the first one, html tags removed
		} else {
			return "", fmt.Errorf("no such element with selector: %s", selector)
		}
	} else {
		return "", err
	}
}

func extractStrings(doc *html.HtmlDocument, selector string) ([]string, error) {
	strs := []string{}
	xpath := css.Convert(selector, css.GLOBAL)

	if searched, err := doc.Search(xpath); err == nil {
		if len(searched) > 0 {
			for _, s := range searched {
				strs = append(strs, s.Content())
			}
			return strs, nil // return all elements, html tags removed
		} else {
			return strs, fmt.Errorf("no such element with selector: %s", selector)
		}
	} else {
		return strs, err
	}
}

// get html attributes for elements with given selector
func extractAttrStrings(doc *html.HtmlDocument, selector, attrName string) (attrs []string, err error) {
	xpath := css.Convert(selector, css.GLOBAL)

	var searched []xml.Node
	if searched, err = doc.Search(xpath); err == nil {
		for _, s := range searched {
			a := s.Attr(attrName)

			attrs = append(attrs, a)
		}

		if len(attrs) <= 0 {
			return attrs, fmt.Errorf("no such element with selector: %s, attrname: %s", selector, attrName)
		}
	}

	return attrs, err
}

// get html attribute for the first element with given selector
func extractFirstAttrString(doc *html.HtmlDocument, selector, attrName string) (attr string, err error) {
	xpath := css.Convert(selector, css.GLOBAL)

	var searched []xml.Node
	if searched, err = doc.Search(xpath); err == nil {
		if len(searched) > 0 {
			attr = searched[0].Attr(attrName)
		} else {
			err = fmt.Errorf("no such attribute: %s with selector: %s", attrName, selector)
		}
	}

	return attr, err
}
