declare module 'sm-crypto' {
  export interface SM2KeyPair {
    publicKey: string
    privateKey: string
  }
  export const sm2: {
    generateKeyPairHex(): SM2KeyPair
    doEncrypt(msg: string, publicKey: string, cipherMode?: number): string
    doDecrypt(encryptData: string, privateKey: string, cipherMode?: number): string
    doSignature(msg: string, privateKey: string, options?: { hash?: boolean }): string
    doVerifySignature(msg: string, signHex: string, publicKey: string, options?: { hash?: boolean }): boolean
  }
  export const sm3: {
    (input: string): string
  }
  export const sm4: {
    encrypt(inArray: number[], key: number[]): number[]
    decrypt(inArray: number[], key: number[]): number[]
    encryptHex(inHex: string, keyHex: string): string
    decryptHex(inHex: string, keyHex: string): string
  }
}
