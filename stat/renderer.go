package stat

import (
	"bytes"
	"fmt"
	"html/template"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"github.com/nfnt/resize"
)

const (
	SampleHtmlTemplate = `<html>
	<head>
		<title>Overwatch: Stats of {{.BattleTag}} / {{.Region}} ({{.Platform}})</title>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
		<meta name="viewport" content="user-scalable=yes, initial-scale=1.0, maximum-scale=1.0, minimum-scale=1.0, width=device-width">
		<style>
			body {
				display: block;
				padding: 3px;
				margin: 3px;
				width: 100%;
				max-width: 100%;
				overflow-x: hidden;
				height: 100%;
				background-color: #405275;
				color: #f0edf2;
				font-family: Futura,century gothic,arial,sans-serif;
			}
			h1,h2,h3,h4,h5 {
				font-family: Koverwatch, sans-serif;
				
				margin-top: 12px;
				margin-bottom: 8px;
			}
			h1 {
			    padding-left: 3px;
			}
			h2 {
			    padding-left: 6px;
			}
			h3 {
			    padding-left: 9px;
			}
			h4 {
			    padding-left: 12px;
			}
			h5 {
			    padding-left: 15px;
			}
			div.info-items {
				overflow-x: auto;
				overflow-y: hidden;
				height: 100px;
				width: 100%;
			}
			div.info-item {
				display: block;
				float: left;
			}
			div.info-name {
				font-size: 1.8rem;
				font-family: Koverwatch, sans-serif;

				margin-bottom: 10px;
			}
			div.info-detail {
				font-family: Koverwatch, sans-serif;

				margin-bottom: 5px;
			}
			img.profile {
				width: 80px;
				display: inline-block;
				vertical-align: middle;
			}
			div.level {
				display: inline-block;
				vertical-align: middle;
				width: 80px;
				height: 80px;
				text-align: center;
				position: relative;
				float: left;
				background-image: url({{.LevelImageUrl}}), url({{.LevelStarImageUrl}});
				background-size: 80px 80px, 80px 40px;
				background-repeat: no-repeat, no-repeat;
				background-position: center, left bottom;
			}
			div.level-text {
				position: relative;
				top: 40%;
				width: 100%;
				font-size: 0.9rem;
			}
			img.level {
				width: 80px;
			}
			div.rank {
				display: inline-block;
				width: 80px;
				height: 80px;
				text-align: center;
				position: relative;
				float: left;
			}
			div.rank.text {
				left: 0;
				position: absolute;
				top: 54px;
				width: 100%;
				font-size: 0.9rem;
				font-size: 0.9rem;
			}
			img.rank {
				width: 60px;
			}
			img.hero-portrait {
				width: 32px;
			}
			img.achievement-icon {
				width: 32px;
			}
			ul {
				list-style-type: none;
				display: table;

				padding-left: 10px;
			}
			ul > li {
				display: table-row;
			}
			ul > li > span {
				display: table-cell;
				padding: 3px 5px 3px 5px;
			}
			li.not-achieved {
				opacity: 0.3;
			}
			span.key {
				color: rgba(240,237,242,.6);
			}
			span.key > img {
				vertical-align: middle;
			}
			span.value {
				font-weight: bold;
			}
		</style>
	</head>
	<body>
		<div id="info">
			<div class="info-name">{{.Name}}</div>
			<div class="info-detail">{{.Detail}}</div>
			<div class="info-items">
				<div class="info-item">
					<img src="{{.ProfileImageUrl}}" class="profile">
				</div>
				<div class="info-item">
					<div class="level">
						<div class="level-text">{{.Level}}</div>
					</div>
				</div>
				{{if .CompetitiveRankImageUrl}}
					<div class="info-item">
						<div class="rank">
							<img src="{{.CompetitiveRankImageUrl}}" class="rank">
							<div class="rank text">{{.CompetitiveRank}}</div>
						</div>
					</div>
				{{end}}
			</div>
		</div>
		<div id="quick-play">
			<h1>Quick Play</h1>
			<div class="featured-stats">
				<h2>Featured Stats</h2>
				<ul>
					{{range $key, $val := .QuickPlay.FeaturedStats}}
						<li>
							<span class="key">{{$key}}</span><span class="value">{{$val}}</span>
						</li>
					{{end}}
				</ul>
			</div>
			<div class="top-heroes">
				<h2>Top Heroes</h2>
				<ul>
				{{range $key, $val := .QuickPlay.TopHeroes}}
					<li>
						<h3>{{$key}}</h3>
						<ul>
							{{range $key, $val := $val}}
								<li>
									<span class="key"><img src="{{$val.ImageUrl}}" class="hero-portrait"> {{$val.Name}}</span><span class="value">{{$val.Value}}</span>
								</li>
							{{end}}
						</ul>
					</li>
				{{end}}
				</ul>
			</div>
			<div class="career-stats">
				<h2>Career Stats</h2>
				<ul>
					{{range .QuickPlay.CareerStats}}
						<li>
							<h3>{{.HeroName}}</h3>
							<ul>
								{{range .Categories}}
									<li>
										<h4>{{.Name}}</h4>
										<ul>
											{{range $key, $val := .Values}}
												<li>
													<span class="key">{{$key}}</span><span class="value">{{$val}}</span>
												</li>
											{{end}}
										</ul>
									</li>
								{{end}}
							</ul>
						</li>
					{{end}}
				</ul>
			</div>
		</div>
		<div id="competitive-play">
			<h1>Competitive Play</h1>
			<div class="featured-stats">
				<h2>Featured Stats</h2>
				<ul>
					{{range $key, $val := .CompetitivePlay.FeaturedStats}}
						<li>
							<span class="key">{{$key}}</span><span class="value">{{$val}}</span>
						</li>
					{{end}}
				</ul>
			</div>
			<div class="top-heroes">
				<h2>Top Heroes</h2>
				<ul>
					{{range $key, $val := .CompetitivePlay.TopHeroes}}
						<li>
							<h3>{{$key}}</h3>
							<ul>
								{{range $key, $val := $val}}
									<li>
										<span class="key"><img src="{{$val.ImageUrl}}" class="hero-portrait"> {{$val.Name}}</span><span class="value">{{$val.Value}}</span>
									</li>
								{{end}}
							</ul>
						</li>
					{{end}}
				</ul>
			</div>
			<div class="career-stats">
				<h2>Career Stats</h2>
				<ul>
					{{range .CompetitivePlay.CareerStats}}
						<li>
							<h3>{{.HeroName}}</h3>
							<ul>
								{{range .Categories}}
									<li>
										<h4>{{.Name}}</h4>
										<ul>
											{{range $key, $val := .Values}}
												<li>
													<span class="key">{{$key}}</span><span class="value">{{$val}}</span>
												</li>
											{{end}}
										</ul>
									</li>
								{{end}}
							</ul>
						</li>
					{{end}}
				</ul>
			</div>
		</div>
		<div id="achievements">
			<h1>Achievements</h1>
			<ul>
			{{range .Achievements}}
				<li>
					<h2>{{.Name}}</h2>
					{{if .Achieved}}
						<ul>
							{{range .Achieved}}
								<li>
									<span class="key"><img src="{{.ImageUrl}}" class="achievement-icon"> {{.Title}}</span><span class="value">{{.Description}}</span>
								</li>
							{{end}}
						</ul>
					{{end}}
					{{if .NonAchieved}}
						<ul>
							{{range .NonAchieved}}
								<li class="not-achieved">
									<span class="key"><img src="{{.ImageUrl}}" class="achievement-icon"> {{.Title}}</span><span class="value">{{.Description}}</span>
								</li>
							{{end}}
						</ul>
					{{end}}
				</li>
			{{end}}
			</ul>
		</div>
	</body>
</html>`
)

// sizes
const (
	// positions and sizes on banner
	BannerWidth        = 320
	BannerHeight       = 50
	Margin             = 4
	BannerRankIconSize = 35
	BannerLevelBgSize  = BannerHeight

	// font sizes
	FontSizeBattleTag float64 = 17.0
	FontSizeDetail    float64 = 13.0
	FontSizeLevel     float64 = 11.0
	FontSizeRank      float64 = 11.0
)

const (
	// file urls
	OverwatchLogoImageUrl = "https://github.com/meinside/overwatch-go/raw/master/overwatch_logo.png"
	KoverwatchFontUrl     = "http://kr.battle.net/forums/static/fonts/koverwatch/koverwatch.ttf"
)

// render given stat to .html format, using template
func RenderStatToHtml(stat Stat, templateStr string) (result string, err error) {
	var tmpl *template.Template
	if tmpl, err = template.New("html").Parse(templateStr); err == nil {
		var buffer bytes.Buffer
		if err = tmpl.Execute(&buffer, stat); err == nil {
			return buffer.String(), nil
		}
	}
	return "", err
}

// render given stat to a banner file in .png format
//
// - when logo is nil: it will be loaded from OverwatchLogoImageUrl
// - when font is nil: it will be loaded from KoverwatchFontUrl
func RenderStatToPngFile(stat Stat, logo image.Image, font *truetype.Font, outFilepath string) error {
	if image, err := genBanner(stat, logo, font); err == nil {
		var file *os.File
		if file, err = os.OpenFile(outFilepath, os.O_WRONLY|os.O_CREATE, 0640); err == nil {
			defer file.Close()

			return png.Encode(file, image)
		} else {
			return err
		}
	} else {
		return err
	}
}

// return bytes of generated banner in .png format
//
// - when logo is nil: it will be loaded from OverwatchLogoImageUrl
// - when font is nil: it will be loaded from KoverwatchFontUrl
func RenderStatToPngBytes(stat Stat, logo image.Image, font *truetype.Font) ([]byte, error) {
	if image, err := genBanner(stat, logo, font); err == nil {
		imgBytes := new(bytes.Buffer)
		if err := png.Encode(imgBytes, image); err == nil {
			return imgBytes.Bytes(), nil
		} else {
			return []byte{}, err
		}
	} else {
		return []byte{}, err
	}
}

// generate a banner image
func genBanner(stat Stat, logo image.Image, font *truetype.Font) (result *image.RGBA, err error) {
	banner := image.NewRGBA(image.Rect(0, 0, BannerWidth, BannerHeight))

	// fill background color (#405275)
	draw.Draw(
		banner,
		banner.Bounds(),
		&image.Uniform{color.RGBA{64, 82, 117, 255}},
		image.ZP,
		draw.Src,
	)

	// load logo image
	if logo == nil {
		if logo, err = getImage(OverwatchLogoImageUrl); err != nil {
			return nil, err
		}
	}

	// resize logo image to fit in the banner
	logo = resize.Resize(BannerHeight, BannerHeight, logo, resize.Lanczos3)

	// draw logo image
	draw.Draw(
		banner,
		image.Rectangle{
			Min: image.Point{X: 0, Y: 0},
			Max: image.Point{X: BannerHeight, Y: BannerHeight},
		},
		logo,
		image.ZP,
		draw.Over,
	)

	// load profile image from url
	var profile image.Image
	if profile, err = getImage(stat.ProfileImageUrl); err != nil {
		return nil, err
	}

	// resize profile image to fit in the banner
	profile = resize.Resize(BannerHeight, BannerHeight, profile, resize.Lanczos3)

	// draw profile image
	draw.Draw(
		banner,
		image.Rectangle{
			Min: image.Point{X: BannerWidth - BannerHeight, Y: 0},
			Max: image.Point{X: BannerWidth, Y: BannerHeight},
		},
		profile,
		image.ZP,
		draw.Over,
	)

	// load .ttf font
	if font == nil {
		if font, err = getFont(KoverwatchFontUrl); err != nil {
			return nil, err
		}
	}

	// setup context
	context := freetype.NewContext()
	context.SetFont(font)
	context.SetDPI(72)
	context.SetClip(banner.Bounds())
	context.SetDst(banner)
	context.SetSrc(image.White)

	// print battletag, platform, and region
	context.SetFontSize(FontSizeBattleTag)
	if _, err = context.DrawString(
		fmt.Sprintf("%s  %s/%s", stat.BattleTag, stat.Platform, stat.Region),
		freetype.Pt(
			int(BannerHeight+Margin),
			int(context.PointToFixed(FontSizeBattleTag)>>6),
		),
	); err != nil {
		return nil, err
	}

	// print detail,
	context.SetFontSize(FontSizeDetail)
	if _, err = context.DrawString(
		stat.Detail,
		freetype.Pt(
			BannerHeight+Margin,
			int(context.PointToFixed(BannerHeight*0.88)>>6),
		),
	); err != nil {
		return nil, err
	}
	// level stars (if exists)
	levelBgStartY := 0.0
	levelTextY := BannerHeight * 0.58
	if stat.LevelStarImageUrl != "" {
		levelBgStartY = -BannerHeight * 0.1
		levelTextY = BannerHeight * 0.48

		var levelStar image.Image
		if levelStar, err = getImage(stat.LevelStarImageUrl); err == nil {
			// load and resize stars to fit in the banner
			levelStar = resize.Resize(BannerLevelBgSize, BannerLevelBgSize*0.5, levelStar, resize.Lanczos3)

			draw.Draw(
				banner,
				image.Rectangle{
					Min: image.Point{X: BannerWidth - BannerHeight*2, Y: BannerLevelBgSize * 0.5},
					Max: image.Point{X: BannerWidth - BannerHeight, Y: BannerLevelBgSize},
				},
				levelStar,
				image.ZP,
				draw.Over,
			)
		} else {
			return nil, err
		}
	}
	// level and level bg
	var levelBg image.Image
	if levelBg, err = getImage(stat.LevelImageUrl); err == nil {
		// load and resize level bg to fit in the banner
		levelBg = resize.Resize(BannerLevelBgSize, BannerLevelBgSize, levelBg, resize.Lanczos3)

		draw.Draw(
			banner,
			image.Rectangle{
				Min: image.Point{X: BannerWidth - BannerHeight*2, Y: int(levelBgStartY)},
				Max: image.Point{X: BannerWidth - BannerHeight, Y: BannerHeight + int(levelBgStartY)},
			},
			levelBg,
			image.ZP,
			draw.Over,
		)
	} else {
		return nil, err
	}
	context.SetFontSize(FontSizeLevel)
	if _, err = context.DrawString(
		fmt.Sprintf("%3d", stat.Level),
		freetype.Pt(
			int(BannerWidth-BannerHeight*1.64),
			int(context.PointToFixed(levelTextY)>>6),
		),
	); err != nil {
		return nil, err
	}
	// rank (only when it exists)
	if stat.CompetitiveRank != NoCompetitiveRank {
		var rankIcon image.Image
		if rankIcon, err = getImage(stat.CompetitiveRankImageUrl); err == nil {
			// resize it to fit in the banner
			rankIcon = resize.Resize(BannerRankIconSize, BannerRankIconSize, rankIcon, resize.Lanczos3)

			draw.Draw(
				banner,
				image.Rectangle{
					Min: image.Point{X: BannerWidth - BannerHeight*2 - BannerRankIconSize, Y: 0},
					Max: image.Point{X: BannerWidth - BannerHeight*2, Y: BannerRankIconSize},
				},
				rankIcon,
				image.ZP,
				draw.Over,
			)
		} else {
			return nil, err
		}
		context.SetFontSize(FontSizeRank)
		if _, err = context.DrawString(
			fmt.Sprintf("%4d", stat.CompetitiveRank),
			freetype.Pt(
				int(BannerWidth-BannerHeight*2.52),
				int(context.PointToFixed(BannerHeight*0.86)>>6),
			),
		); err != nil {
			return nil, err
		}
	}
	return banner, nil
}

// read image from given url
func getImage(url string) (image.Image, error) {
	if res, err := http.Get(url); err == nil {
		defer res.Body.Close()

		if img, _, err := image.Decode(res.Body); err == nil {
			return img, nil
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

// read ttf font from given url
func getFont(url string) (*truetype.Font, error) {
	if res, err := http.Get(url); err == nil {
		defer res.Body.Close()

		if bytes, err := ioutil.ReadAll(res.Body); err == nil {
			if font, err := truetype.Parse(bytes); err == nil {
				return font, nil
			} else {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}
