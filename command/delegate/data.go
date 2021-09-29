package delegate

import (
	"encoding/json"
	"io"
)

func getJSONDataFromReader(r io.Reader, data interface{}) error {
	err := json.NewDecoder(r).Decode(data)
	if err != nil {
		return err
	}

	return nil
}

func GetSetupRequest(r io.Reader) (*SetupRequest, error) {
	d := &SetupRequest{}
	if err := getJSONDataFromReader(r, d); err != nil {
		return nil, err
	}

	if d.StageEnvStr != "" {
		var env map[string]string
		if err := json.Unmarshal([]byte(d.StageEnvStr), &env); err != nil {
			return nil, err
		}
		d.StageEnvVars = env
	}

	if d.SecretEnvStr != "" {
		var secretEnv map[string]string
		if err := json.Unmarshal([]byte(d.SecretEnvStr), &secretEnv); err != nil {
			return nil, err
		}
		d.SecretEnvVars = secretEnv
	}

	return d, nil
}

type SetupRequest struct {
	StageID string `json:"stage_id"`

	StageEnvStr  string `json:"stage_env"`
	StageEnvVars map[string]string

	SecretEnvStr  string `json:"secret_env"`
	SecretEnvVars map[string]string

	DataDump SetupDataDump `json:"dump"`
}

type SetupDataDump struct {
	Class             string      `json:"@class"`
	DelegateSelectors interface{} `json:"delegateSelectors"`
	K8SConnector      struct {
		ConnectorConfig struct {
			Credential struct {
				Type string `json:"type"`
				Spec struct {
					MasterUrl string `json:"masterUrl"`
					Auth      struct {
						Type string `json:"type"`
						Spec struct {
							Username    string      `json:"username"`
							UsernameRef interface{} `json:"usernameRef"`
							PasswordRef string      `json:"passwordRef"`
						} `json:"spec"`
					} `json:"auth"`
				} `json:"spec"`
			} `json:"credential"`
			DelegateSelectors []interface{} `json:"delegateSelectors"`
		} `json:"connectorConfig"`
		ConnectorType        string      `json:"connectorType"`
		Identifier           string      `json:"identifier"`
		OrgIdentifier        interface{} `json:"orgIdentifier"`
		ProjectIdentifier    interface{} `json:"projectIdentifier"`
		DelegateSelectors    interface{} `json:"delegateSelectors"`
		EncryptedDataDetails []struct {
			EncryptedData struct {
				Uuid                 string      `json:"uuid"`
				Name                 string      `json:"name"`
				Path                 interface{} `json:"path"`
				Parameters           interface{} `json:"parameters"`
				EncryptionKey        string      `json:"encryptionKey"`
				EncryptedValue       string      `json:"encryptedValue"`
				KmsId                interface{} `json:"kmsId"`
				EncryptionType       string      `json:"encryptionType"`
				BackupEncryptedValue interface{} `json:"backupEncryptedValue"`
				BackupEncryptionKey  interface{} `json:"backupEncryptionKey"`
				BackupKmsId          interface{} `json:"backupKmsId"`
				BackupEncryptionType interface{} `json:"backupEncryptionType"`
				Base64Encoded        bool        `json:"base64Encoded"`
				AdditionalMetadata   interface{} `json:"additionalMetadata"`
			} `json:"encryptedData"`
			EncryptionConfig struct {
				Name                                          interface{} `json:"name"`
				Uuid                                          interface{} `json:"uuid"`
				EncryptionType                                string      `json:"encryptionType"`
				AccountId                                     string      `json:"accountId"`
				NumOfEncryptedValue                           int         `json:"numOfEncryptedValue"`
				EncryptedBy                                   interface{} `json:"encryptedBy"`
				CreatedBy                                     interface{} `json:"createdBy"`
				CreatedAt                                     int         `json:"createdAt"`
				LastUpdatedBy                                 interface{} `json:"lastUpdatedBy"`
				LastUpdatedAt                                 int         `json:"lastUpdatedAt"`
				NextTokenRenewIteration                       interface{} `json:"nextTokenRenewIteration"`
				ManuallyEnteredSecretEngineMigrationIteration interface{} `json:"manuallyEnteredSecretEngineMigrationIteration"`
				UsageRestrictions                             interface{} `json:"usageRestrictions"`
				ScopedToAccount                               bool        `json:"scopedToAccount"`
				TemplatizedFields                             interface{} `json:"templatizedFields"`
				ValidationCriteria                            string      `json:"validationCriteria"`
				EncryptionServiceUrl                          interface{} `json:"encryptionServiceUrl"`
				Default                                       bool        `json:"default"`
				Templatized                                   bool        `json:"templatized"`
			} `json:"encryptionConfig"`
			FieldName  string `json:"fieldName"`
			Identifier struct {
				KmsId         interface{} `json:"kmsId"`
				EncryptionKey string      `json:"encryptionKey"`
			} `json:"identifier"`
		} `json:"encryptedDataDetails"`
		SshKeyDetails   interface{} `json:"sshKeyDetails"`
		EnvToSecretsMap struct {
		} `json:"envToSecretsMap"`
	} `json:"k8sConnector"`
	Cik8PodParams struct {
		Name        string      `json:"name"`
		Namespace   string      `json:"namespace"`
		Annotations interface{} `json:"annotations"`
		Labels      struct {
			AccountID           string `json:"accountID"`
			PipelineExecutionID string `json:"pipelineExecutionID"`
			ProjectID           string `json:"projectID"`
			OrgID               string `json:"orgID"`
			PipelineID          string `json:"pipelineID"`
			StageID             string `json:"stageID"`
		} `json:"labels"`
		ContainerParamsList []struct {
			Name                      string `json:"name"`
			ImageDetailsWithConnector struct {
				ImageConnectorDetails *struct {
					ConnectorConfig struct {
						DockerRegistryUrl string `json:"dockerRegistryUrl"`
						ProviderType      string `json:"providerType"`
						Auth              struct {
							Type string `json:"type"`
						} `json:"auth"`
						DelegateSelectors []interface{} `json:"delegateSelectors"`
					} `json:"connectorConfig"`
					ConnectorType        string      `json:"connectorType"`
					Identifier           string      `json:"identifier"`
					OrgIdentifier        interface{} `json:"orgIdentifier"`
					ProjectIdentifier    interface{} `json:"projectIdentifier"`
					DelegateSelectors    interface{} `json:"delegateSelectors"`
					EncryptedDataDetails interface{} `json:"encryptedDataDetails"`
					SshKeyDetails        interface{} `json:"sshKeyDetails"`
					EnvToSecretsMap      struct {
					} `json:"envToSecretsMap"`
				} `json:"imageConnectorDetails"`
				ImageDetails struct {
					Name        string      `json:"name"`
					Tag         string      `json:"tag"`
					SourceName  interface{} `json:"sourceName"`
					RegistryUrl interface{} `json:"registryUrl"`
					Username    interface{} `json:"username"`
					UsernameRef interface{} `json:"usernameRef"`
					Password    interface{} `json:"password"`
					DomainName  interface{} `json:"domainName"`
				} `json:"imageDetails"`
			} `json:"imageDetailsWithConnector"`
			Commands             []string          `json:"commands"`
			Args                 []string          `json:"args"`
			WorkingDir           string            `json:"workingDir"`
			Ports                []int             `json:"ports"`
			EnvVars              map[string]string `json:"envVars"`
			EnvVarsWithSecretRef *struct {
			} `json:"envVarsWithSecretRef"`
			SecretEnvVars     interface{} `json:"secretEnvVars"`
			SecretVolumes     interface{} `json:"secretVolumes"`
			ImageSecret       interface{} `json:"imageSecret"`
			VolumeToMountPath struct {
				Addon   string `json:"addon"`
				Harness string `json:"harness"`
			} `json:"volumeToMountPath"`
			ContainerResourceParams struct {
				ResourceRequestMemoryMiB int `json:"resourceRequestMemoryMiB"`
				ResourceLimitMemoryMiB   int `json:"resourceLimitMemoryMiB"`
				ResourceRequestMilliCpu  int `json:"resourceRequestMilliCpu"`
				ResourceLimitMilliCpu    int `json:"resourceLimitMilliCpu"`
			} `json:"containerResourceParams"`
			ContainerSecrets struct {
				SecretVariableDetails []interface{} `json:"secretVariableDetails"`
				ConnectorDetailsMap   struct {
				} `json:"connectorDetailsMap"`
				FunctorConnectors struct {
				} `json:"functorConnectors"`
				PlainTextSecretsByName map[string]PlainTextSecretByName `json:"plainTextSecretsByName"`
			} `json:"containerSecrets"`
			RunAsUser       interface{} `json:"runAsUser"`
			Privileged      bool        `json:"privileged"`
			ImagePullPolicy interface{} `json:"imagePullPolicy"`
			ContainerType   string      `json:"containerType"`
			Type            string      `json:"type"`
		} `json:"containerParamsList"`
		InitContainerParamsList []struct {
			Name                      string `json:"name"`
			ImageDetailsWithConnector struct {
				ImageConnectorDetails interface{} `json:"imageConnectorDetails"`
				ImageDetails          struct {
					Name        string      `json:"name"`
					Tag         string      `json:"tag"`
					SourceName  interface{} `json:"sourceName"`
					RegistryUrl interface{} `json:"registryUrl"`
					Username    interface{} `json:"username"`
					UsernameRef interface{} `json:"usernameRef"`
					Password    interface{} `json:"password"`
					DomainName  interface{} `json:"domainName"`
				} `json:"imageDetails"`
			} `json:"imageDetailsWithConnector"`
			Commands   []string    `json:"commands"`
			Args       []string    `json:"args"`
			WorkingDir interface{} `json:"workingDir"`
			Ports      interface{} `json:"ports"`
			EnvVars    struct {
				HARNESSWORKSPACE string `json:"HARNESS_WORKSPACE"`
			} `json:"envVars"`
			EnvVarsWithSecretRef interface{} `json:"envVarsWithSecretRef"`
			SecretEnvVars        interface{} `json:"secretEnvVars"`
			SecretVolumes        interface{} `json:"secretVolumes"`
			ImageSecret          interface{} `json:"imageSecret"`
			VolumeToMountPath    struct {
				Addon   string `json:"addon"`
				Harness string `json:"harness"`
			} `json:"volumeToMountPath"`
			ContainerResourceParams interface{} `json:"containerResourceParams"`
			ContainerSecrets        struct {
				SecretVariableDetails []interface{} `json:"secretVariableDetails"`
				ConnectorDetailsMap   struct {
				} `json:"connectorDetailsMap"`
				FunctorConnectors struct {
				} `json:"functorConnectors"`
				PlainTextSecretsByName struct {
				} `json:"plainTextSecretsByName"`
			} `json:"containerSecrets"`
			RunAsUser       interface{} `json:"runAsUser"`
			Privileged      bool        `json:"privileged"`
			ImagePullPolicy interface{} `json:"imagePullPolicy"`
			ContainerType   string      `json:"containerType"`
			Type            string      `json:"type"`
		} `json:"initContainerParamsList"`
		PvcParamList        []interface{} `json:"pvcParamList"`
		HostAliasParamsList interface{}   `json:"hostAliasParamsList"`
		RunAsUser           interface{}   `json:"runAsUser"`
		ServiceAccountName  interface{}   `json:"serviceAccountName"`
		GitConnector        interface{}   `json:"gitConnector"`
		BranchName          interface{}   `json:"branchName"`
		CommitId            interface{}   `json:"commitId"`
		StepExecVolumeName  interface{}   `json:"stepExecVolumeName"`
		StepExecWorkingDir  interface{}   `json:"stepExecWorkingDir"`
		Type                string        `json:"type"`
	} `json:"cik8PodParams"`
	ServicePodParams         interface{} `json:"servicePodParams"`
	PodMaxWaitUntilReadySecs int         `json:"podMaxWaitUntilReadySecs"`
	Type                     string      `json:"type"`
}

type PlainTextSecretByName struct {
	Value     string `json:"value"`
	SecretKey string `json:"secretKey"`
	Type      string `json:"type"`
}

func (d *SetupRequest) GetStageID() string {
	return d.StageID
	//return d.DataDump.Cik8PodParams.Name
}

func (d *SetupRequest) StepCount() int {
	return len(d.DataDump.Cik8PodParams.ContainerParamsList)
}

func (d *SetupRequest) StepName(i int) string {
	return d.DataDump.Cik8PodParams.ContainerParamsList[i].Name
}

func (d *SetupRequest) StepImage(i int) string {
	return d.DataDump.Cik8PodParams.ContainerParamsList[i].ImageDetailsWithConnector.ImageDetails.Name + ":" +
		d.DataDump.Cik8PodParams.ContainerParamsList[i].ImageDetailsWithConnector.ImageDetails.Tag
}

func (d *SetupRequest) StepEnv(i int) map[string]string {
	return d.DataDump.Cik8PodParams.ContainerParamsList[i].EnvVars
}

func GetDestroyRequest(r io.Reader) (*DestroyRequest, error) {
	d := &DestroyRequest{}
	if err := getJSONDataFromReader(r, d); err != nil {
		return nil, err
	}

	return d, nil
}

type DestroyRequest struct {
	StageID string `json:"stage_id"`
}

func GetExecStepRequest(r io.Reader) (*ExecStepRequest, error) {
	d := &ExecStepRequest{}
	if err := getJSONDataFromReader(r, d); err != nil {
		return nil, err
	}

	if d.EnvStr != "" {
		var env map[string]string
		if err := json.Unmarshal([]byte(d.EnvStr), &env); err != nil {
			return nil, err
		}
		d.EnvVars = env
	}

	return d, nil
}

type ExecStepRequest struct {
	StageID            string `json:"stage_id"`
	StepID             string `json:"step_id"`
	Command            string `json:"command"`
	Image              string `json:"image"`
	LogKey             string `json:"log_key"`
	LogStreamURL       string `json:"log_stream_url"`
	LogStreamAccountID string `json:"log_stream_account_id"`
	LogStreamToken     string `json:"log_stream_token"`

	EnvStr  string `json:"env"`
	EnvVars map[string]string

	Dump ExecStepDataDump `json:"dump"`
}

type ExecStepDataDump struct {
	ExecutionId string `json:"executionId"`
	Step        struct {
		Id          string `json:"id"`
		DisplayName string `json:"displayName"`
		Run         struct {
			Command string `json:"command"`
			Context struct {
				NumRetries           int    `json:"numRetries"`
				ExecutionTimeoutSecs string `json:"executionTimeoutSecs"`
			} `json:"context"`
			EnvVarOutputs []interface{} `json:"envVarOutputs"`
			ContainerPort int           `json:"containerPort"`
			Reports       []interface{} `json:"reports"`
			Environment   struct {
			} `json:"environment"`
			ShellType string `json:"shellType"`
			Image     string `json:"image"`
		} `json:"run"`
		CallbackToken string `json:"callbackToken"`
		TaskId        string `json:"taskId"`
		SkipCondition string `json:"skipCondition"`
		LogKey        string `json:"logKey"`
		AccountId     string `json:"accountId"`
		ContainerPort int    `json:"containerPort"`
	} `json:"step"`
	TmpFilePath         string `json:"tmpFilePath"`
	DelegateSvcEndpoint string `json:"delegateSvcEndpoint"`
}
