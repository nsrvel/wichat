package response

// ErrorBody is the client-facing error payload (i18n code + interpolation params).
type ErrorBody struct {
	Code   string         `json:"code"`
	Params map[string]any `json:"params"`
}

// Envelope wraps API error responses.
type Envelope struct {
	Error ErrorBody `json:"error"`
}

// NewErrorEnvelope builds a JSON-safe error envelope.
func NewErrorEnvelope(code string, params map[string]any) Envelope {
	return Envelope{Error: ErrorBody{Code: code, Params: params}}
}

// DataEnvelope is the client-facing success payload.
type DataEnvelope struct {
	Data any `json:"data"`
}

// NewDataEnvelope wraps a successful response body.
func NewDataEnvelope(data any) DataEnvelope {
	return DataEnvelope{Data: data}
}
