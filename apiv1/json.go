package apiv1

import "encoding/json"

type jsonEncoder struct{}

func (j jsonEncoder) encode(resp Response) ([]byte, error) {
	return json.Marshal(resp)
}

func (j jsonEncoder) contentType() string {
	return "application/json"
}
