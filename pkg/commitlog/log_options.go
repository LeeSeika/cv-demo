package commitlog

type LogOptions struct {
	logDir          string
	logMaxBytes     uint64
	indexMaxBytes   uint64
	messageMaxBytes uint64
}

func DefaultLogOptions(dir string) LogOptions {
	return LogOptions{
		logDir:          dir,
		logMaxBytes:     100000000,
		indexMaxBytes:   800000,
		messageMaxBytes: 1000000,
	}
}

func (o *LogOptions) SetIndexMaxItems(items uint64) {
	o.indexMaxBytes = items * IndexEntryBytes
}

func (o *LogOptions) SetSegmentMaxBytes(maxBytes uint64) {
	o.logMaxBytes = maxBytes
}

func (o *LogOptions) SetMessageMaxBytes(maxBytes uint64) {
	o.messageMaxBytes = maxBytes
}
