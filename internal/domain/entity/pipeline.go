package entity

// UnifiedPipelineDefinition is the top-level struct for the entire YAML file.
type UnifiedPipelineDefinition struct {
	Pipeline  Pipeline   `yaml:"pipeline"`
	Resources *Resources `yaml:"resources,omitempty"`
}

// Pipeline holds the core pipeline metadata and the list of steps.
type Pipeline struct {
	Name        string                 `yaml:"name"`
	Version     string                 `yaml:"version"`
	Domain      string                 `yaml:"domain"`
	Description string                 `yaml:"description"`
	Owners      []Owner                `yaml:"owners,omitempty"`
	Schedule    *Schedule              `yaml:"schedule,omitempty"`
	Parameters  map[string]interface{} `yaml:"parameters,omitempty"` // could be more structured if you prefer
	Steps       []Step                 `yaml:"steps"`
}

type Owner struct {
	Name  string `yaml:"name"`
	Email string `yaml:"email"`
}

type Schedule struct {
	Type       string `yaml:"type,omitempty"`
	Expression string `yaml:"expression,omitempty"`
}

// Step represents a discrete stage in the pipeline (ingestion, transformation, etc.).
type Step struct {
	Name                string         `yaml:"name"`
	Type                string         `yaml:"type"`
	Description         string         `yaml:"description"`
	DependsOn           []string       `yaml:"depends_on,omitempty"`
	Inputs              []DataRef      `yaml:"inputs,omitempty"`
	Outputs             []DataRef      `yaml:"outputs,omitempty"`
	Config              interface{}    `yaml:"config,omitempty"` // or map[string]interface{}
	TransformationQuery string         `yaml:"transformation_query,omitempty"`
	Notifications       *Notifications `yaml:"notifications,omitempty"`
}

// DataRef captures how a step references input/output data (e.g., S3 paths, table names).
type DataRef struct {
	Name      string `yaml:"name"`
	Type      string `yaml:"type"`
	Path      string `yaml:"path,omitempty"`       // e.g., s3 bucket path
	TableName string `yaml:"table_name,omitempty"` // e.g., staging.customer_activity
	Host      string `yaml:"host,omitempty"`       // e.g., db host
	Database  string `yaml:"database,omitempty"`
	// Add more fields if needed (e.g., secrets, authentication, etc.)
}

// Notifications encapsulates the actions to be taken on success or failure.
type Notifications struct {
	OnSuccess []NotificationTarget `yaml:"on_success,omitempty"`
	OnFailure []NotificationTarget `yaml:"on_failure,omitempty"`
}

// NotificationTarget describes where and how a notification is sent.
type NotificationTarget struct {
	Method     string   `yaml:"method"`
	Recipients []string `yaml:"recipients,omitempty"`
	Channel    string   `yaml:"channel,omitempty"`
	// Add more fields depending on your notification channels.
}

// Resources define optional platform-level resources (compute, storage, etc.).
type Resources struct {
	ComputeCluster  string `yaml:"compute_cluster,omitempty"`
	StorageLocation string `yaml:"storage_location,omitempty"`
}
