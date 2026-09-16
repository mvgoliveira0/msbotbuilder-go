package schema

// UploadInfo contains info to upload contents
type UploadInfo struct {
	// ContentUrl is a direct link to the final location of the file on OneDrive.
	ContentURL string
	// FileType is the file type, as determined by OneDrive.
	FileType string
	// Name is the name of the file. Note that this may be different from the name that the bot proposed initially.
	Name string
	// UniqueID is an unique ID set for the contents.
	UniqueID string
	// UploadURL is the upload URL of the file. This points to a OneDrive upload session for the file. The upload session is valid for 15 minutes.
	UploadURL string
}
