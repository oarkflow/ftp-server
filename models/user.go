package models

type User struct {
	ID          int64    `json:"id"`
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	Permissions []string `json:"permissions"`
}

// TypeCredential is the type of credential used for authentication.
type TypeCredential string

const (
	Password  TypeCredential = "PASSWORD"
	TwoFactor TypeCredential = "TWO_FACTOR"
	APIKey    TypeCredential = "API_Key"
	Oauth     TypeCredential = "OAUTH"
)

type Integration string

const (
	FTP  Integration = "FTP"
	SFTP Integration = "SFTP"
	Mail Integration = "MAIL"
)

// TypeProvider is the type of provider used for authentication.
type TypeProvider string

const (
	Local   TypeProvider = "LOCAL"
	Cognito TypeProvider = "COGNITO"
	Dropbox TypeProvider = "DROPBOX"
	GDrive  TypeProvider = "GDRIVE"
	S3      TypeProvider = "S3"
)

type Credential struct {
	Credential     string         `json:"credential"`
	CredentialType TypeCredential `json:"credential_type"`
	ProviderType   TypeProvider   `json:"provider_type"`
	Integration    Integration    `json:"integration"`
	CredentialID   int64          `gorm:"primaryKey" json:"credential_id"`
	UserID         int64          `json:"user_id"`
}
