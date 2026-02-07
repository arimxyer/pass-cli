package recovery

// Default KDF Parameters (Argon2id)
const (
	DefaultTime    uint32 = 1     // Single pass (memory cost primary defense)
	DefaultMemory  uint32 = 65536 // 64 MB (RFC 9106 recommended default)
	DefaultThreads uint8  = 4     // Matches existing vault KDF, utilizes modern CPUs
	DefaultKeyLen  uint32 = 32    // AES-256
	DefaultSaltLen int    = 32    // 256 bits
)

// BIP39 Constants
const (
	EntropyBits    int = 256  // 24-word mnemonic
	MnemonicWords  int = 24   // Fixed word count
	WordlistSize   int = 2048 // Standard BIP39 wordlist size
	ChallengeCount int = 6    // Fixed: 6-word challenge
	VerifyCount    int = 3    // Default: 3-word verification
)

// AES-GCM Constants
const (
	GCMNonceSize int = 12 // Standard GCM nonce size (96 bits)
)
