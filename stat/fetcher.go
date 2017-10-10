package stat

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

const (
	NoCompetitiveRank = -1
)

const (
	PlatformPc  = "pc"
	PlatformXbl = "xbl"
	PlatformPsn = "psn"
)

var Verbose bool = false

// generate url for given params
//
// ex:
//		https://playoverwatch.com/en-us/career/pc/kr/meinside-3155
//		https://playoverwatch.com/ko-kr/career/xbl/meinside
//		https://playoverwatch.com/ru-ru/career/psn/meinside
func GenUrl(battleTagString string, battleTagNumber int, platform, region, language string) string {
	if strings.EqualFold(platform, PlatformPc) {
		return fmt.Sprintf("https://playoverwatch.com/%s/career/%s/%s/%s-%d",
			language,
			platform,
			region,
			battleTagString,
			battleTagNumber,
		)
	} else {
		return fmt.Sprintf("https://playoverwatch.com/%s/career/%s/%s",
			language,
			platform,
			battleTagString,
		)
	}
}

// fetch given user's stat from official overwatch site.
func FetchStat(battleTagString string, battleTagNumber int, platform, region, language string) (result Stat, err error) {
	url := GenUrl(battleTagString, battleTagNumber, platform, region, language)

	if Verbose {
		log.Printf("> fetching from url: %s\n", url)
	}

	// fetch html document,
	var doc *goquery.Document
	doc, err = goquery.NewDocument(url)
	if err != nil {
		return Stat{}, err
	}

	if Verbose {
		log.Printf("> received html document: %+v\n", doc)
	}

	// parse it and assign to struct
	return parseStat(doc, battleTagString, battleTagNumber, platform, region)
}

// parse stat from html bytes
//
// XXX - if it stops working, should check the html response and alter css selectors
func parseStat(doc *goquery.Document, battleTagString string, battleTagNumber int, platform, region string) (result Stat, err error) {
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
	if level, err = extractInt32(doc, "div.player-level > div:nth-child(1)"); err != nil {
		return Stat{}, err
	}
	var levelImageUrl string
	if levelImageUrl, err = extractFirstAttrString(doc, "div.player-level", "style"); err != nil {
		return Stat{}, err
	} else {
		// XXX - strip background-image:url(...)
		if strings.HasPrefix(levelImageUrl, "background-image:url(") {
			levelImageUrl = strings.TrimLeft(levelImageUrl, "background-image:url(")
		}
		if strings.HasSuffix(levelImageUrl, ")") {
			levelImageUrl = strings.TrimRight(levelImageUrl, ")")
		}
	}
	var levelStarImageUrl string
	if levelStarImageUrl, err = extractFirstAttrString(doc, "div.player-level > div.player-rank", "style"); err == nil {
		// XXX - strip background-image:url(...)
		if strings.HasPrefix(levelStarImageUrl, "background-image:url(") {
			levelStarImageUrl = strings.TrimLeft(levelStarImageUrl, "background-image:url(")
		}
		if strings.HasSuffix(levelStarImageUrl, ")") {
			levelStarImageUrl = strings.TrimRight(levelStarImageUrl, ")")
		}
	}
	var competitiveRank int32
	if competitiveRank, err = extractInt32(doc, "div.competitive-rank > div"); err != nil {
		competitiveRank = NoCompetitiveRank
	}
	var competitiveRankImageUrl string
	competitiveRankImageUrl, _ = extractFirstAttrString(doc, "div.competitive-rank > img", "src")
	var detail string
	if detail, err = extractString(doc, "div.masthead > p.masthead-detail > span"); err != nil {
		return Stat{}, err
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
	// [achievements]
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
		if urls, err = extractAttrStrings(doc, fmt.Sprintf("#achievements-section > div > div:nth-of-type(%d) > ul div.achievement-card > img", i+2 /* skip first one */), "src"); err != nil {
			return Stat{}, err
		}
		if titles, err = extractStrings(doc, fmt.Sprintf("#achievements-section > div > div:nth-of-type(%d) div.tooltip-tip > h6", i+2 /* skip first one */)); err != nil {
			return Stat{}, err
		}
		if descriptions, err = extractStrings(doc, fmt.Sprintf("#achievements-section > div > div:nth-of-type(%d) div.tooltip-tip > p", i+2 /* skip first one */)); err != nil {
			return Stat{}, err
		}
		if classes, err = extractAttrStrings(doc, fmt.Sprintf("#achievements-section > div > div:nth-of-type(%d) > ul div.achievement-card", i+2 /* skip first one */), "class"); err != nil {
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

	var battleTag string
	if strings.EqualFold(platform, PlatformPc) {
		battleTag = fmt.Sprintf("%s#%d", battleTagString, battleTagNumber)
	} else {
		battleTag = battleTagString
	}

	// return result
	return Stat{
		BattleTag: battleTag,
		Platform:  platform,
		Region:    region,

		Name:                    name,
		ProfileImageUrl:         profileImageUrl,
		Level:                   level,
		LevelImageUrl:           levelImageUrl,
		LevelStarImageUrl:       levelStarImageUrl,
		CompetitiveRank:         competitiveRank,
		CompetitiveRankImageUrl: competitiveRankImageUrl,
		Detail:                  detail,

		QuickPlay:       quickPlayStat,
		CompetitivePlay: competitivePlayStat,

		Achievements: achievements,
	}, err
}

func extractPlayStat(doc *goquery.Document, id TagId) (featuredStats map[string]string, topHeroes map[string][]Hero, careerStats []CareerStat, err error) {
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
		if heroNames, err = extractStrings(doc, fmt.Sprintf("#%s > section.hero-comparison-section > div > div:nth-of-type(%d) div.bar-text > div.title", id, i+2 /* skip first one */)); err != nil {
			return featuredStats, topHeroes, careerStats, err
		}
		if heroImageUrls, err = extractAttrStrings(doc, fmt.Sprintf("#%s > section.hero-comparison-section > div > div:nth-of-type(%d) img", id, i+2 /* skip first one */), "src"); err != nil {
			return featuredStats, topHeroes, careerStats, err
		}
		if heroValues, err = extractStrings(doc, fmt.Sprintf("#%s > section.hero-comparison-section > div > div:nth-of-type(%d) div.bar-text > div.description", id, i+2 /* skip first one */)); err != nil {
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
		if categoryNames, err = extractStrings(doc, fmt.Sprintf("#%s div[data-category-id=\"%s\"] div.card-stat-block > table.data-table > thead > tr > th > .stat-title", id, statId)); err != nil {
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

func extractInt32(doc *goquery.Document, selector string) (int32, error) {
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

func extractInt64(doc *goquery.Document, selector string) (int64, error) {
	if s, err := extractString(doc, selector); err == nil {
		return strconv.ParseInt(strings.Replace(s, ",", "", -1), 10, 64) // XXX - remove unwanted ','
	} else {
		return 0, err
	}
}

func extractFloat32(doc *goquery.Document, selector string) (float32, error) {
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

func extractString(doc *goquery.Document, selector string) (string, error) {
	var result string
	exists := false

	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		result = s.Text()
		exists = true
		return
	})

	if exists {
		return result, nil
	} else {
		return "", fmt.Errorf("no such element with selector: %s", selector)
	}
}

func extractStrings(doc *goquery.Document, selector string) ([]string, error) {
	strs := []string{}

	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		strs = append(strs, s.Text())
	})

	if len(strs) > 0 {
		return strs, nil // return all elements, html tags removed
	} else {
		return strs, fmt.Errorf("no such element with selector: %s", selector)
	}
}

// get html attributes for elements with given selector
func extractAttrStrings(doc *goquery.Document, selector, attrName string) ([]string, error) {
	attrs := []string{}

	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		if attr, exists := s.Attr(attrName); exists {
			attrs = append(attrs, attr)
		}
	})

	if len(attrs) > 0 {
		return attrs, nil
	} else {
		return []string{}, fmt.Errorf("no such attr with selector: %s, attrname: %s", selector, attrName)
	}
}

// get html attribute for the first element with given selector
func extractFirstAttrString(doc *goquery.Document, selector, attrName string) (string, error) {
	var attr string
	exists := false

	doc.Find(selector).Each(func(i int, s *goquery.Selection) {
		if attr, exists = s.Attr(attrName); exists {
			return
		}
	})

	if exists {
		return attr, nil
	} else {
		return "", fmt.Errorf("no such attr with selector: %s, attrname: %s", selector, attrName)
	}
}
