package main

type TableEntryExact struct {
	TableName  string            `json:"tableName"`
	ActionName string            `json:"actionName"`
	Params     []string          `json:"params"`
	MatchField map[string]string `json:"matchField"`
}
