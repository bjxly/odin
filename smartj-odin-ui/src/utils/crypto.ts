import { sm2 } from 'sm-crypto'

/**
 * SM2 公钥加密
 * @param plaintext 明文
 * @param publicKey SM2 公钥（hex 格式）
 * @returns 加密后的密文（hex 格式）
 */
export function encryptBySM2(plaintext: string, publicKey: string): string {
  // cipherMode: 1 = C1C3C2 标准模式
  return sm2.doEncrypt(plaintext, publicKey, 1)
}

/**
 * 生成 SM2 密钥对
 * 通常由后端生成，前端仅用于测试
 */
export function generateSM2KeyPair(): { publicKey: string; privateKey: string } {
  const keypair = sm2.generateKeyPairHex()
  return {
    publicKey: keypair.publicKey,
    privateKey: keypair.privateKey,
  }
}
