package stat

type TagId string

const (
	TagIdQuickPlay       TagId = "quick-play"
	TagIdCompetitivePlay TagId = "competitive-play"
)

// stat struct fetched from official site
type Stat struct {
	// info
	Name            string `json:"name"`
	ProfileImageUrl string `json:"profile_image_url"`
	Level           int32  `json:"level"`
	CompetitiveRank int32  `json:"competitive_rank"`

	// stats: quick play
	QuickPlay PlayStat `json:"quick_play"`

	// stats: competitive play
	CompetitivePlay PlayStat `json:"competitive_play"`

	// achievements
	Achievements []AchievementCategory `json:"achievements"`
}

type PlayStat struct {
	// featured stats
	FeaturedStats map[string]string `json:"featured_stats"`

	// top heroes
	TopHeroes map[string][]Hero `json:"top_heroes"`

	// career stats
	CareerStats []CareerStat `json:"career_stats"`
}

type Hero struct {
	Name     string `json:"name"`
	ImageUrl string `json:"image_url"`
	Value    string `json:"value"`
}

type CareerStat struct {
	HeroName   string               `json:"hero_name"`
	Categories []CareerStatCategory `json:"categories"`
}

type CareerStatCategory struct {
	Name   string            `json:"name"`
	Values map[string]string `json:"values"`
}

type AchievementCategory struct {
	Name        string        `json:"name"`
	Achieved    []Achievement `json:"achieved"`
	NonAchieved []Achievement `json:"non_achieved"`
}

type Achievement struct {
	Title       string `json:"title"`
	Description string `json:"achievement"`
	ImageUrl    string `json:"image_url"`
}
