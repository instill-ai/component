package googlecloudstorage

type ReadInput struct {
	BucketName               string
	Delimiter                string
	Prefix                   string
	Versions                 bool
	StartOffset              string
	EndOffset                string
	IncludeTrailingDelimiter bool
	MatchGlob                string
	IncludeFoldersAsPrefixes bool
}

type ReadOutput struct {
	TextObjects     []TextObject
	ImageObjects    []ImageObject
	DocumentObjects []DocumentObject
}

type TextObject struct {
	Data       string
	Attributes Attributes
}

type ImageObject struct {
	Data       string
	Attributes Attributes
}

type DocumentObject struct {
	Data       string
	Attributes Attributes
}

type Attributes struct {
	Name               string
	ContentType        string
	ContentLanguage    string
	Owner              string
	Size               int64
	ContentEncoding    string
	ContentDisposition string
	MD5                []byte
	MediaLink          string
	Metadata           map[string]string
	StorageClass       string
}

