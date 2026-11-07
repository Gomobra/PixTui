package qrcode

import (
	"github.com/Gomobra/PixTui/packages/solid"
)

// RegisterQRCodeSolid registers the QR code renderable with the Solid
// component registry under the "qr_code" name.
func RegisterQRCodeSolid() {
	solid.RegisterComponent("qr_code", func(props map[string]interface{}) solid.SolidComponent {
		text, _ := props["text"].(string)
		qr, err := NewQRCode(text, LevelM)
		if err != nil {
			return solid.SolidComponent{
				Type: "text",
				Props: map[string]interface{}{
					"content": err.Error(),
				},
			}
		}
		return solid.SolidComponent{
			Type: "box",
			Props: map[string]interface{}{
				"content": qr.Render(),
			},
		}
	})
}
