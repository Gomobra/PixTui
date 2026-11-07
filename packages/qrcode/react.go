package qrcode

import (
	"github.com/Gomobra/PixTui/packages/react"
)

// RegisterQRCodeReact registers the QR code renderable with the React
// component registry under the "qr-code" name.
func RegisterQRCodeReact() {
	react.RegisterComponent("qr-code", func(props map[string]interface{}) react.ReactElement {
		text, _ := props["text"].(string)
		qr, err := NewQRCode(text, LevelM)
		if err != nil {
			return react.ReactElement{
				Type: "text",
				Props: map[string]interface{}{
					"content": err.Error(),
				},
			}
		}
		return react.ReactElement{
			Type: "box",
			Props: map[string]interface{}{
				"content": qr.Render(),
			},
		}
	})
}
