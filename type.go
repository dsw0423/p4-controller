package main

type TableEntryExact struct {
	TableName  string            `json:"tableName"`
	ActionName string            `json:"actionName"`
	Params     []string          `json:"params"`
	MatchField map[string]string `json:"matchField"`
}

type FileNameHash struct {
	FileName string `json:"fileName"`
	Hash     string `json:"hash"`
}

type FileInfo struct {
	FileName  string `json:"fileName" redis:"fileName"`
	Size      int64  `json:"size" redis:"size"`
	Timestamp int64  `json:"timestamp" redis:"timestamp"`
}

type FileInfoString struct {
	FileName  string `json:"fileName"`
	Size      string `json:"size"`
	Timestamp string `json:"timestamp"`
	Hash      string `json:"hash"`
}
