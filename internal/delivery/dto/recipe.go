package dto

import "encoding/json"

type RecipeDto struct {
	Id          int             `json:"id,omitempty"`
	Name        string          `json:"name,omitempty"`
	Desc        string          `json:"description,omitempty"`
	Img         string          `json:"img,omitempty"`
	CookingTime int             `json:"cookingTime,omitempty"`
	ServingsNum int             `json:"servingsNum,omitempty"`
	Ingredients string          `json:"ingredients,omitempty"`
	Steps       json.RawMessage `json:"steps,omitempty"`
}
