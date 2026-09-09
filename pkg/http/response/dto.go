package response

type Err struct {
	Code int    `json:"code"`
	Text string `json:"text"`
}

type Resp struct {
	Error    *Err `json:"error,omitempty"`
	Response any  `json:"response,omitempty"`
	Data     any  `json:"data,omitempty"`
}
