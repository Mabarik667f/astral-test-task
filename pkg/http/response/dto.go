package response

type Err struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

type Resp struct {
	Error    Err            `json:"error,omitzero"`
	Response map[string]any `json:"response,omitempty"`
	Data     map[string]any `json:"data,omitempty"`
}
