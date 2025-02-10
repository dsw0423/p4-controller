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

type Ports struct {
	IDs []int `json:"/ethdev/list"`
}

type PortInfo struct {
	PortId  int    `json:"portId"`
	MacAddr string `json:"mac_addr"`
	Mtu     int    `json:"mtu"`
}

type PortInfoRoot struct {
	PortInfo `json:"/ethdev/info"`
}

type PortStatus struct {
	PortId int    `json:"portId"`
	Status string `json:"status"`
	Speed  int    `json:"speed"`
	Duplex string `json:"duplex"`
}

type PortStatusRoot struct {
	PortStatus `json:"/ethdev/link_status"`
}

type PortStats struct {
	PortId    int    `json:"portId"`
	RxPackets uint64 `json:"ipackets"`
	TxPackets uint64 `json:"opackets"`
	RxBytes   uint64 `json:"ibytes"`
	TxBytes   uint64 `json:"obytes"`
}

type PortStatsRoot struct {
	PortStats `json:"/ethdev/stats"`
}

type PortBitRate struct {
	PortId int     `json:"portId"`
	RxRate float64 `json:"rxRate"`
	TxRate float64 `json:"txRate"`
}
