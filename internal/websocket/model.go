package websocket

type BalanceUpdate struct {
	MemberID  int64   `json:"member_id"`
	CurrentID int64   `json:"currency_id"`
	Balance   float64 `json:"balance"`
	Action    string  `json:"action"`
	Topic     string  `json:"topic"`
}
type BetSattlement struct {
	Message string `json:"message"`
	Time    string `json:"time"`
	Topic   string `json:"topic"`
}
type BalanceUpdateMessage struct {
	Pattern string `json:"pattern"`
	Data    string `json:"data"`
}

type Data struct {
	ID            int     `json:"id"`
	MemberLoginID string  `json:"member_login_id"`
	MemberID      int     `json:"member_id"`
	CurrencyID    int     `json:"currency_id"`
	Balance       float64 `json:"balance"`
}
