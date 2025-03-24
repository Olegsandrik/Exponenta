package utils

import "fmt"

const zeroFilter = `{ "match_all": {} }`

func FilterForElasticsearchRecipeIndex(maxTime int, dishType string, diet string) (string, string, string) {
	maxTimeFilter := zeroFilter

	if maxTime != 0 {
		maxTimeFilter = fmt.Sprintf(`{ "range": { "cookingTime": { "gte": 0, "lte": %d }}}`, maxTime)
	}

	dishTypeFilter := zeroFilter

	if dishType != "" {
		dishTypeFilter = fmt.Sprintf(` { "match": { "dishTypes": "%s" } }`, dishType)
	}

	dietFilter := zeroFilter

	if diet != "" {
		dietFilter = fmt.Sprintf(`{ "match": { "diets": "%s" } }`, diet)
	}

	return maxTimeFilter, dishTypeFilter, dietFilter
}
