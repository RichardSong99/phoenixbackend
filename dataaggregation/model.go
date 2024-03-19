package dataaggregation

type StatusStat struct {
	Correct     float64 `bson:"correct"`
	Flagged     float64 `bson:"flagged"`
	Incorrect   float64 `bson:"incorrect"`
	Omitted     float64 `bson:"omitted"`
	Total       float64 `bson:"total"`
	Unattempted float64 `bson:"unattempted"`
}

type TopicStat struct {
	Easy    *StatusStat `bson:"easy,omitempty"`
	Medium  *StatusStat `bson:"medium,omitempty"`
	Hard    *StatusStat `bson:"hard,omitempty"`
	Extreme *StatusStat `bson:"extreme,omitempty"`
	Total   *StatusStat `bson:"total,omitempty"`
	Topic   string      `bson:"topic"`
}

//==========================================================

type DifficultyAggregation struct {
	Easy    *int `json:"easy,omitempty"`
	Medium  *int `json:"medium,omitempty"`
	Hard    *int `json:"hard,omitempty"`
	Extreme *int `json:"extreme,omitempty"`
}

type StatusAggregation struct {
	Difficulties DifficultyAggregation `json:"difficulties"`
}

type TopicAggregation struct {
	Statuses map[string]StatusAggregation `json:"statuses"`
	Topic    string                       `json:"topic"`
}

type Topics []TopicAggregation
