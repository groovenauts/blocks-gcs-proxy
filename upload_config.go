package main

type UploadConfig struct {
	Worker           *WorkerConfig `json:"worker,omitempty"`
	ContentTypeByExt bool          `json:"content_type_by_ext,omitempty"`
}
