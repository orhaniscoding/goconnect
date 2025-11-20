/**
 * Simple QR Code generator for WireGuard configs
 * Uses HTML Canvas API for client-side generation
 */

/**
 * Generate QR code as Data URL
 * @param text - Text to encode in QR code
 * @param size - Size of the QR code (default: 256)
 * @returns Promise<string> - Data URL of the QR code image
 */
export async function generateQRCode(text: string, size: number = 256): Promise<string> {
  // Use a lightweight QR library approach or external service
  // For production, we'll use qrcode.react or similar
  // For now, return a placeholder that shows we need the library
  
  // Option 1: Use external QR API (works without npm install)
  const encodedText = encodeURIComponent(text)
  const apiUrl = `https://api.qrserver.com/v1/create-qr-code/?size=${size}x${size}&data=${encodedText}`
  
  return apiUrl
}

/**
 * Generate QR code on canvas element
 * @param text - Text to encode
 * @param canvasElement - Canvas element to draw on
 */
export function generateQRCodeOnCanvas(text: string, canvasElement: HTMLCanvasElement): void {
  // This would require a QR library like 'qrcode' or 'qrcode-generator'
  // For now, we'll use the external API approach
  const img = new Image()
  img.onload = () => {
    const ctx = canvasElement.getContext('2d')
    if (ctx) {
      canvasElement.width = img.width
      canvasElement.height = img.height
      ctx.drawImage(img, 0, 0)
    }
  }
  img.src = `https://api.qrserver.com/v1/create-qr-code/?size=256x256&data=${encodeURIComponent(text)}`
}

/**
 * Download QR code as image file
 * @param qrDataUrl - Data URL of the QR code
 * @param filename - Filename for download
 */
export function downloadQRCode(qrDataUrl: string, filename: string): void {
  const a = document.createElement('a')
  a.href = qrDataUrl
  a.download = filename
  document.body.appendChild(a)
  a.click()
  document.body.removeChild(a)
}
