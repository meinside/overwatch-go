package stat

import (
	"bytes"
	"html/template"
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
				font-size: 3rem;
				font-family: Koverwatch, sans-serif;
			}
			div.info-detail {
				font-family: Koverwatch, sans-serif;
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
			}
			div.level.text {
				left: 0;
				position: absolute;
				top: 32px;
				width: 100%;
				font-size: 0.9rem;
			}
			img.level {
				width: 80;
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
			}
			ul > li {
				display: table-row;
			}
			ul > li > span {
				display: table-cell;
				padding: 3px 5px 3px 5px;
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
						<img src="{{.LevelImageUrl}}" class="level">
						<div class="level text">{{.Level}}</div>
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
						<h3>Achieved</h3>
						<ul>
							{{range .Achieved}}
								<li>
									<span class="key"><img src="{{.ImageUrl}}" class="achievement-icon"> {{.Title}}</span><span class="value">{{.Description}}</span>
								</li>
							{{end}}
						</ul>
					{{end}}
					{{if .NonAchieved}}
						<h3>Not achieved yet</h3>
						<ul>
							{{range .NonAchieved}}
								<li>
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

// render given stat to html format, using template
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
