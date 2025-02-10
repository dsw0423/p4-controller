package main

import (
	"bytes"
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
)

func notPrimary() bool {
	return !isPrimary
}

func stringToByteSlice(s string) []byte {
	ss := strings.Split(s, ":")
	data := make([]byte, len(ss))
	for i, b := range ss {
		d, _ := strconv.ParseUint(b, 16, 8)
		data[i] = byte(d)
	}
	return data
}

func byteSliceToString(bytes []byte) string {
	var builder strings.Builder
	for _, b := range bytes {
		builder.WriteString(strconv.FormatUint(uint64(b), 16))
		builder.WriteByte(byte(':'))
	}
	res := builder.String()
	res = res[:len(res)-1]
	return res
}

func getPorts() *Ports {
	inputData := `/ethdev/list`
	cmd := exec.Command("python3", "/home/dsw/codes/p4-controller/dpdk-telemetry.py")
	cmd.Stdin = bytes.NewBufferString(inputData)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()

	var ports Ports
	json.Unmarshal(out.Bytes(), &ports)
	return &ports
}

func getPortInfo(portId int) *PortInfo {
	inputData := `/ethdev/info,` + strconv.Itoa(portId)
	cmd := exec.Command("python3", "/home/dsw/codes/p4-controller/dpdk-telemetry.py")
	cmd.Stdin = bytes.NewBufferString(inputData)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()

	var portInfo PortInfo
	var portInfoRoot PortInfoRoot

	json.Unmarshal(out.Bytes(), &portInfoRoot)
	portInfo.PortId = portId
	portInfo.MacAddr = portInfoRoot.MacAddr
	portInfo.Mtu = portInfoRoot.Mtu

	return &portInfo
}

func getPortStatus(portId int) *PortStatus {
	inputData := `/ethdev/link_status,` + strconv.Itoa(portId)
	cmd := exec.Command("python3", "/home/dsw/codes/p4-controller/dpdk-telemetry.py")
	cmd.Stdin = bytes.NewBufferString(inputData)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()

	var portStatus PortStatus
	var portStatusRoot PortStatusRoot

	json.Unmarshal(out.Bytes(), &portStatusRoot)
	portStatus.PortId = portId
	portStatus.Status = portStatusRoot.Status
	portStatus.Speed = portStatusRoot.Speed
	portStatus.Duplex = portStatusRoot.Duplex

	return &portStatus
}

func getPortStats(portId int) *PortStats {
	inputData := `/ethdev/stats,` + strconv.Itoa(portId)
	cmd := exec.Command("python3", "/home/dsw/codes/p4-controller/dpdk-telemetry.py")
	cmd.Stdin = bytes.NewBufferString(inputData)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run()

	var portStats PortStats
	var portStatsRoot PortStatsRoot

	json.Unmarshal(out.Bytes(), &portStatsRoot)
	portStats.PortId = portId
	portStats.RxPackets = portStatsRoot.RxPackets
	portStats.TxPackets = portStatsRoot.TxPackets
	portStats.RxBytes = portStatsRoot.RxBytes
	portStats.TxBytes = portStatsRoot.TxBytes

	return &portStats
}
