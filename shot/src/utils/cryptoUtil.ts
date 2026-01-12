// AES-GCM加密解密工具类
class CryptoUtil {
    // 与服务器端保持一致的密钥字符串
    private static readonly KEY_STRING = '32字节密钥1234567890123456'
    
    // 获取32字节密钥
    private static get KEY(): Uint8Array {
        const encoded = new TextEncoder().encode(this.KEY_STRING)
        const key = new Uint8Array(32)
        
        // 与后端逻辑保持一致：如果密钥长度大于等于32字节，截断为32字节；否则填充到32字节
        if (encoded.length >= 32) {
            key.set(encoded.slice(0, 32))
        } else {
            key.set(encoded)
        }
        
        return key
    }
    
    // 辅助方法：将ArrayBuffer转换为Base64字符串（兼容后端base64.StdEncoding）
    private static arrayBufferToBase64(buffer: ArrayBuffer): string {
        const bytes = new Uint8Array(buffer)
        let binary = ''
        for (let i = 0; i < bytes.byteLength; i++) {
            binary += String.fromCharCode(bytes[i])
        }
        return btoa(binary)
    }
    
    // 辅助方法：将Base64字符串转换为ArrayBuffer（兼容后端base64.StdEncoding）
    private static base64ToArrayBuffer(base64: string): ArrayBuffer {
        const binary = atob(base64)
        const bytes = new Uint8Array(binary.length)
        for (let i = 0; i < binary.length; i++) {
            bytes[i] = binary.charCodeAt(i)
        }
        return bytes.buffer
    }
    
    // 加密函数
    static async encrypt(plaintext: string): Promise<string> {
        // 生成随机IV（12字节）
        const iv = crypto.getRandomValues(new Uint8Array(12))
        
        // 导入密钥
        const cryptoKey = await crypto.subtle.importKey(
            'raw',
            this.KEY.buffer as ArrayBuffer,
            { name: 'AES-GCM' },
            false,
            ['encrypt', 'decrypt']
        )
        
        // 加密数据
        const encodedData = new TextEncoder().encode(plaintext)
        const encrypted = await crypto.subtle.encrypt(
            {
                name: 'AES-GCM',
                iv: iv
            },
            cryptoKey,
            encodedData
        )
        
        // 合并IV和密文
        const result = new Uint8Array(iv.length + encrypted.byteLength)
        result.set(iv)
        result.set(new Uint8Array(encrypted), iv.length)
        
        // Base64编码，确保与后端base64.StdEncoding兼容
        return this.arrayBufferToBase64(result.buffer)
    }
    
    // 解密函数
    static async decrypt(ciphertext: string): Promise<string> {
        // Base64解码，确保与后端base64.StdEncoding兼容
        const encryptedBuffer = this.base64ToArrayBuffer(ciphertext)
        const encryptedData = new Uint8Array(encryptedBuffer)
        
        // 分离IV和密文
        const iv = encryptedData.slice(0, 12)
        const data = encryptedData.slice(12)
        
        // 导入密钥
        const cryptoKey = await crypto.subtle.importKey(
            'raw',
            this.KEY.buffer as ArrayBuffer,
            { name: 'AES-GCM' },
            false,
            ['encrypt', 'decrypt']
        )
        
        // 解密数据
        const decrypted = await crypto.subtle.decrypt(
            {
                name: 'AES-GCM',
                iv: iv
            },
            cryptoKey,
            data
        )
        
        // 解码为字符串
        return new TextDecoder().decode(decrypted)
    }
}

export default CryptoUtil