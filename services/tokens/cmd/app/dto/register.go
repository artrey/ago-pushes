package dto

type RegisterToken struct {
	UserID    int64  `json:"userId"`
	PushToken string `json:"pushToken"`
}
