package smp

// CBOR request/response payloads for SMP, as in Zephyr v3.7.0.
// Each type links to the corresponding payload definition in the spec:
// https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_protocol.html

// Generic error fields present in all responses.
type MinimalResponse struct {
	Error     *ErrorResponse `cbor:"err,omitempty" json:"error,omitempty"`
	ErrorCode *uint8         `cbor:"rc,omitempty" json:"errorCode,omitempty"`
}

// SMP v2 structured error response.
type ErrorResponse struct {
	Group Group `cbor:"group" json:"group"`
	Code  uint8 `cbor:"rc" json:"code"`
}

// ============================================================================
// Group 0 — OS Management
// ============================================================================

// OSEchoRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_0.html#echo-request
type OSEchoRequest struct {
	Data string `cbor:"d"`
}

// OSEchoResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_0.html#echo-response
type OSEchoResponse struct {
	MinimalResponse
	Output string `cbor:"r" json:"output"`
}

// OSTaskStatsRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_0.html#task-statistics-request
type OSTaskStatsRequest struct{}

// OSTaskStatsResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_0.html#task-statistics-response
type OSTaskStatsResponse struct {
	MinimalResponse
	Tasks map[string]OSTaskStat `cbor:"tasks" json:"tasks"`
}

// OSTaskStat: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_0.html#task-statistics-response
type OSTaskStat struct {
	Priority        uint64 `cbor:"prio" json:"priority"`
	TaskID          uint64 `cbor:"tid" json:"taskID"`
	State           uint64 `cbor:"state" json:"state"`
	StackUsage      uint64 `cbor:"stkuse" json:"stackUsage"`
	StackSize       uint64 `cbor:"stksiz" json:"stackSize"`
	ContextSwitches uint64 `cbor:"cswcnt" json:"contextSwitches"`
	Runtime         uint64 `cbor:"runtime" json:"runtime"`
	LastCheckin     uint64 `cbor:"last_checkin" json:"lastCheckin"`
	NextCheckin     uint64 `cbor:"next_checkin" json:"nextCheckin"`
}

// NB: OSMempoolStats breaks the standard model since its payload is keyed at the top level
/*
type OSMempoolStatsRequest struct{}

type OSMempoolStatsResponse map[string]OSMempoolStat

type OSMempoolStat struct {
	BlockSize int64 `cbor:"blksiz"`
	Blocks    int64 `cbor:"nblks"`
	Free      int64 `cbor:"nfree"`
	Min       int64 `cbor:"min"`
}
*/

// OSDatetimeReadRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_0.html#date-time-get-request
type OSDatetimeReadRequest struct{}

// OSDatetimeReadResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_0.html#date-time-get-response
type OSDatetimeReadResponse struct {
	MinimalResponse
	Datetime string `cbor:"datetime" json:"datetime"`
}

// OSDatetimeWriteRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_0.html#date-time-set-request
type OSDatetimeWriteRequest struct {
	Datetime string `cbor:"datetime"`
}

// OSDatetimeWriteResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_0.html#date-time-set-response
type OSDatetimeWriteResponse struct {
	MinimalResponse
}

// OSResetRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_0.html#system-reset-request
type OSResetRequest struct {
	Force *uint8 `cbor:"force,omitempty"`
}

// OSResetResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_0.html#system-reset-response
type OSResetResponse struct {
	MinimalResponse
}

// OSMcuMgrParametersRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_0.html#mcumgr-parameters-request
type OSMcuMgrParametersRequest struct{}

// OSMcuMgrParametersResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_0.html#mcumgr-parameters-response
type OSMcuMgrParametersResponse struct {
	MinimalResponse
	BufSize  uint64 `cbor:"buf_size" json:"bufSize"`
	BufCount uint64 `cbor:"buf_count" json:"bufCount"`
}

// OSAppInfoRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_0.html#os-application-info-request
type OSAppInfoRequest struct {
	Format *string `cbor:"format,omitempty"`
}

// OSAppInfoResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_0.html#os-application-info-response
type OSAppInfoResponse struct {
	MinimalResponse
	Output string `cbor:"output" json:"output"`
}

// OSBootloaderInfoRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_0.html#bootloader-information-request
type OSBootloaderInfoRequest struct {
	Query *string `cbor:"query,omitempty"`
}

// OSBootloaderInfoResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_0.html#bootloader-information-response
type OSBootloaderInfoResponse struct {
	MinimalResponse
	Bootloader  *string `cbor:"bootloader" json:"bootloader,omitempty"`
	Mode        *int    `cbor:"mode" json:"mode,omitempty"`
	NoDowngrade *bool   `cbor:"no-downgrade" json:"noDowngrade,omitempty"`
}

// ============================================================================
// Group 1 — Image Management
// ============================================================================

// ImageStateRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_1.html#set-state-of-image-request
type ImageStateRequest struct {
	Confirm bool   `cbor:"confirm"`
	Hash    []byte `cbor:"hash,omitempty"`
}

// ImageStateResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_1.html#get-state-of-images-response
type ImageStateResponse struct {
	MinimalResponse
	Images      []ImageStateResponseImage `cbor:"images" json:"images"`
	SplitStatus *uint64                   `cbor:"splitStatus,omitempty" json:"splitStatus,omitempty"`
}

// ImageStateResponseImage: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_1.html#get-state-of-images-response
type ImageStateResponseImage struct {
	Image     *uint8 `cbor:"image,omitempty" json:"image,omitempty"`
	Slot      uint8  `cbor:"slot" json:"slot"`
	Version   string `cbor:"version" json:"version"`
	Hash      []byte `cbor:"hash" json:"hash"`
	Bootable  *bool  `cbor:"bootable,omitempty" json:"bootable,omitempty"`
	Pending   *bool  `cbor:"pending,omitempty" json:"pending,omitempty"`
	Confirmed *bool  `cbor:"confirmed,omitempty" json:"confirmed,omitempty"`
	Active    *bool  `cbor:"active,omitempty" json:"active,omitempty"`
	Permanent *bool  `cbor:"permanent,omitempty" json:"permanent,omitempty"`
}

// ImageUploadRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_1.html#image-upload-request
type ImageUploadRequest struct {
	Length  *uint32 `cbor:"len,omitempty"`
	Data    []byte  `cbor:"data"`
	Offset  uint32  `cbor:"off"`
	Hash    []byte  `cbor:"sha,omitempty"`
	Image   *uint32 `cbor:"image,omitempty"`
	Upgrade *bool   `cbor:"upgrade,omitempty"`
}

// ImageUploadResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_1.html#image-upload-response
type ImageUploadResponse struct {
	MinimalResponse
	Offset uint32 `cbor:"off" json:"offset"`
	Match  *bool  `cbor:"match,omitempty" json:"match,omitempty"`
}

// ImageEraseRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_1.html#image-erase-request
type ImageEraseRequest struct {
	Slot *uint8 `cbor:"slot,omitempty"`
}

// ImageEraseResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_1.html#image-erase-response
type ImageEraseResponse struct {
	MinimalResponse
}

// ============================================================================
// Group 2 — Statistics Management
// ============================================================================

// StatGroupDataRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_2.html#statistics-group-data-request
type StatGroupDataRequest struct {
	Name string `cbor:"name"`
}

// StatGroupDataResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_2.html#statistics-group-data-response
type StatGroupDataResponse struct {
	MinimalResponse
	Name   string            `cbor:"name" json:"name"`
	Fields map[string]uint64 `cbor:"fields" json:"fields"`
}

// StatListGroupsRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_2.html#statistics-list-of-groups-request
type StatListGroupsRequest struct{}

// StatListGroupsResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_2.html#statistics-list-of-groups-response
type StatListGroupsResponse struct {
	MinimalResponse
	StatList []string `cbor:"stat_list" json:"statList"`
}

// ============================================================================
// Group 3 — Settings (Config) Management
// ============================================================================

// SettingsReadRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_3.html#read-setting-request
type SettingsReadRequest struct {
	Name    string  `cbor:"name"`
	MaxSize *uint64 `cbor:"max_size,omitempty"`
}

// SettingsReadResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_3.html#read-setting-response
type SettingsReadResponse struct {
	MinimalResponse
	Value   []byte  `cbor:"val" json:"value"`
	MaxSize *uint64 `cbor:"max_size,omitempty" json:"maxSize,omitempty"`
}

// SettingsWriteRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_3.html#write-setting-request
type SettingsWriteRequest struct {
	Name  string `cbor:"name"`
	Value []byte `cbor:"val"`
}

// SettingsWriteResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_3.html#write-setting-response
type SettingsWriteResponse struct {
	MinimalResponse
}

// SettingsDeleteRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_3.html#delete-setting-request
type SettingsDeleteRequest struct {
	Name string `cbor:"name"`
}

// SettingsDeleteResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_3.html#delete-setting-response
type SettingsDeleteResponse struct {
	MinimalResponse
}

// SettingsCommitRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_3.html#commit-settings-request
type SettingsCommitRequest struct{}

// SettingsCommitResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_3.html#commit-settings-response
type SettingsCommitResponse struct {
	MinimalResponse
}

// SettingsLoadSaveRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_3.html#load-settings-request
type SettingsLoadSaveRequest struct{}

// SettingsLoadSaveResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_3.html#load-settings-response
type SettingsLoadSaveResponse struct {
	MinimalResponse
}

// ============================================================================
// Group 8 — File Management
// ============================================================================

// FsDownloadRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_8.html#file-download-request
type FsDownloadRequest struct {
	Offset uint64 `cbor:"off"`
	Name   string `cbor:"name"`
}

// FsDownloadResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_8.html#file-download-response
type FsDownloadResponse struct {
	MinimalResponse
	Offset uint64  `cbor:"off" json:"offset"`
	Data   []byte  `cbor:"data" json:"data"`
	Length *uint64 `cbor:"len,omitempty" json:"length,omitempty"`
}

// FsUploadRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_8.html#file-upload-request
type FsUploadRequest struct {
	Offset uint64  `cbor:"off"`
	Data   []byte  `cbor:"data"`
	Name   string  `cbor:"name"`
	Length *uint64 `cbor:"len,omitempty"`
}

// FsUploadResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_8.html#file-upload-response
type FsUploadResponse struct {
	MinimalResponse
	Offset uint64 `cbor:"off" json:"offset"`
}

// FsStatusRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_8.html#file-status-request
type FsStatusRequest struct {
	Name string `cbor:"name"`
}

// FsStatusResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_8.html#file-status-response
type FsStatusResponse struct {
	MinimalResponse
	Length uint64 `cbor:"len" json:"length"`
}

// FsHashRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_8.html#file-hash-checksum-request
type FsHashRequest struct {
	Name   string  `cbor:"name"`
	Type   *string `cbor:"type,omitempty"`
	Offset *uint64 `cbor:"off,omitempty"`
	Length *uint64 `cbor:"len,omitempty"`
}

// FsHashResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_8.html#file-hash-checksum-response
type FsHashResponse struct {
	MinimalResponse
	Type   string  `cbor:"type" json:"type"`
	Offset *uint64 `cbor:"off,omitempty" json:"offset,omitempty"`
	Length uint64  `cbor:"len" json:"length"`
	Output any     `cbor:"output" json:"output"` // can be uint (crc) or bytestring (sha256)
}

// FsSupportedHashTypesRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_8.html#supported-file-hash-checksum-types-request
type FsSupportedHashTypesRequest struct{}

// FsSupportedHashTypesResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_8.html#supported-file-hash-checksum-types-response
type FsSupportedHashTypesResponse struct {
	MinimalResponse
	Types map[string]FsHashType `cbor:"types" json:"types"`
}

// FsHashType: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_8.html#hash-checksum-types
type FsHashType struct {
	Format uint64 `cbor:"format" json:"format"`
	Size   uint64 `cbor:"size" json:"size"`
}

// FsCloseRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_8.html#file-close-request
type FsCloseRequest struct{}

// FsCloseResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_8.html#file-close-response
type FsCloseResponse struct {
	MinimalResponse
}

// ============================================================================
// Group 9 — Shell Management
// ============================================================================

// ShellExecRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_9.html#shell-command-line-execute-request
type ShellExecRequest struct {
	Argv []string `cbor:"argv"`
}

// ShellExecResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_9.html#shell-command-line-execute-response
type ShellExecResponse struct {
	MinimalResponse
	Output string `cbor:"o" json:"output"`
	Return int    `cbor:"ret" json:"return"`
}

// ============================================================================
// Group 63 — Zephyr-specific Management
// ============================================================================

// ZephyrEraseRequest: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_63.html#erase-storage-request
type ZephyrEraseRequest struct{}

// ZephyrEraseResponse: https://docs.zephyrproject.org/3.7.0/services/device_mgmt/smp_groups/smp_group_63.html#erase-storage-response
type ZephyrEraseResponse struct {
	MinimalResponse
}
