package domain

type DataManagementPostgresConfig struct {
	Host               string `json:"host"`
	Port               int32  `json:"port"`
	User               string `json:"user"`
	Password           string `json:"password,omitempty"`
	PasswordConfigured bool   `json:"password_configured"`
	Database           string `json:"database"`
	SSLMode            string `json:"ssl_mode"`
	ContainerName      string `json:"container_name"`
}

type DataManagementRedisConfig struct {
	Addr               string `json:"addr"`
	Username           string `json:"username"`
	Password           string `json:"password,omitempty"`
	PasswordConfigured bool   `json:"password_configured"`
	DB                 int32  `json:"db"`
	ContainerName      string `json:"container_name"`
}

type DataManagementS3Config struct {
	Enabled                   bool   `json:"enabled"`
	Endpoint                  string `json:"endpoint"`
	Region                    string `json:"region"`
	Bucket                    string `json:"bucket"`
	AccessKeyID               string `json:"access_key_id"`
	SecretAccessKey           string `json:"secret_access_key,omitempty"`
	SecretAccessKeyConfigured bool   `json:"secret_access_key_configured"`
	Prefix                    string `json:"prefix"`
	ForcePathStyle            bool   `json:"force_path_style"`
	UseSSL                    bool   `json:"use_ssl"`
}

type DataManagementConfig struct {
	SourceMode        string                       `json:"source_mode"`
	BackupRoot        string                       `json:"backup_root"`
	SQLitePath        string                       `json:"sqlite_path,omitempty"`
	RetentionDays     int32                        `json:"retention_days"`
	KeepLast          int32                        `json:"keep_last"`
	ActivePostgresID  string                       `json:"active_postgres_profile_id"`
	ActiveRedisID     string                       `json:"active_redis_profile_id"`
	Postgres          DataManagementPostgresConfig `json:"postgres"`
	Redis             DataManagementRedisConfig    `json:"redis"`
	S3                DataManagementS3Config       `json:"s3"`
	ActiveS3ProfileID string                       `json:"active_s3_profile_id"`
}

type DataManagementTestS3Result struct {
	OK      bool   `json:"ok"`
	Message string `json:"message"`
}

type DataManagementCreateBackupJobInput struct {
	BackupType     string
	UploadToS3     bool
	TriggeredBy    string
	IdempotencyKey string
	S3ProfileID    string
	PostgresID     string
	RedisID        string
}

type DataManagementListBackupJobsInput struct {
	PageSize   int32
	PageToken  string
	Status     string
	BackupType string
}

type DataManagementArtifactInfo struct {
	LocalPath string `json:"local_path"`
	SizeBytes int64  `json:"size_bytes"`
	SHA256    string `json:"sha256"`
}

type DataManagementS3ObjectInfo struct {
	Bucket string `json:"bucket"`
	Key    string `json:"key"`
	ETag   string `json:"etag"`
}

type DataManagementBackupJob struct {
	JobID          string                     `json:"job_id"`
	BackupType     string                     `json:"backup_type"`
	Status         string                     `json:"status"`
	TriggeredBy    string                     `json:"triggered_by"`
	IdempotencyKey string                     `json:"idempotency_key,omitempty"`
	UploadToS3     bool                       `json:"upload_to_s3"`
	S3ProfileID    string                     `json:"s3_profile_id,omitempty"`
	PostgresID     string                     `json:"postgres_profile_id,omitempty"`
	RedisID        string                     `json:"redis_profile_id,omitempty"`
	StartedAt      string                     `json:"started_at,omitempty"`
	FinishedAt     string                     `json:"finished_at,omitempty"`
	ErrorMessage   string                     `json:"error_message,omitempty"`
	Artifact       DataManagementArtifactInfo `json:"artifact"`
	S3Object       DataManagementS3ObjectInfo `json:"s3"`
}

type DataManagementSourceProfile struct {
	SourceType         string                     `json:"source_type"`
	ProfileID          string                     `json:"profile_id"`
	Name               string                     `json:"name"`
	IsActive           bool                       `json:"is_active"`
	Config             DataManagementSourceConfig `json:"config"`
	PasswordConfigured bool                       `json:"password_configured"`
	CreatedAt          string                     `json:"created_at,omitempty"`
	UpdatedAt          string                     `json:"updated_at,omitempty"`
}

type DataManagementSourceConfig struct {
	Host          string `json:"host"`
	Port          int32  `json:"port"`
	User          string `json:"user"`
	Password      string `json:"password,omitempty"`
	Database      string `json:"database"`
	SSLMode       string `json:"ssl_mode"`
	Addr          string `json:"addr"`
	Username      string `json:"username"`
	DB            int32  `json:"db"`
	ContainerName string `json:"container_name"`
}

type DataManagementCreateSourceProfileInput struct {
	SourceType string
	ProfileID  string
	Name       string
	Config     DataManagementSourceConfig
	SetActive  bool
}

type DataManagementUpdateSourceProfileInput struct {
	SourceType string
	ProfileID  string
	Name       string
	Config     DataManagementSourceConfig
}

type DataManagementS3Profile struct {
	ProfileID                 string                 `json:"profile_id"`
	Name                      string                 `json:"name"`
	IsActive                  bool                   `json:"is_active"`
	S3                        DataManagementS3Config `json:"s3"`
	SecretAccessKeyConfigured bool                   `json:"secret_access_key_configured"`
	CreatedAt                 string                 `json:"created_at,omitempty"`
	UpdatedAt                 string                 `json:"updated_at,omitempty"`
}

type DataManagementCreateS3ProfileInput struct {
	ProfileID string
	Name      string
	S3        DataManagementS3Config
	SetActive bool
}

type DataManagementUpdateS3ProfileInput struct {
	ProfileID string
	Name      string
	S3        DataManagementS3Config
}

type DataManagementListBackupJobsResult struct {
	Items         []DataManagementBackupJob `json:"items"`
	NextPageToken string                    `json:"next_page_token,omitempty"`
}
