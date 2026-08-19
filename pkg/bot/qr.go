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

	qrterminal.GenerateHalfBlock(code, qrterminal.L, os.Stdout)

	fmt.Println("==================================================")
	fmt.Println("Buka WhatsApp > Perangkat Tertaut > Tautkan Perangkat")
	fmt.Println("==================================================")
}
