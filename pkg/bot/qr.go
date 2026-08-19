package bot

import (
	"fmt"
	"os"

	"github.com/mdp/qrterminal/v3"
)

// DisplayQRCode prints the QR code in terminal for WhatsApp Web authentication
func DisplayQRCode(code string) {
	fmt.Println("\n==================================================")
	fmt.Println("📱 SCAN QR CODE BERIKUT DENGAN APLIKASI WHATSAPP:")
	fmt.Println("==================================================")

	config := qrterminal.Config{
		Level:     qrterminal.L,
		Writer:    os.Stdout,
		BlackChar: qrterminal.BLACK,
		WhiteChar: qrterminal.WHITE,
		QuietZone: 1,
	}

	qrterminal.GenerateWithConfig(code, config)
	fmt.Println("==================================================")
	fmt.Println("Buka WhatsApp > Perangkat Tertaut > Tautkan Perangkat")
	fmt.Println("==================================================")
}
