package storage

import (
	"sync"
	"time"
)

type (
	UploadRequest struct {
		LocalRoot string       // local root dir（need upload dir）
		RemoteDir string       // Remote target directory on the OSS (e.g. component/2025/itc-testing)
		Files     []FileObject // Automatically generated file list
		Mu        sync.RWMutex // files safe
	}

	FileObject struct {
		LocalPath   string            // Local absolute path (automatically generated)
		RemoteKey   string            // Remote path (auto-generated)
		Content     []byte            // File content (pre-processing loading)
		Metadata    map[string]string // Metadata
		ContentType string            // Clear content type
		Overwrite   bool              // Overwrite existing files
	}

	ObjectInfo struct {
		Key          string    // Object key (full path)
		Size         int64     // File size (bytes)
		LastModified time.Time // Last modified time
		ETag         string    // ETag checksum of the file
		StorageClass string    // Storage Class
		Content      []byte    // File contents (optional)
	}
)
