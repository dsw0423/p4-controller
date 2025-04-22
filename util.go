package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func notPrimaryControllerFor(host *HostInfo) bool {
	return !host.IsPrimary
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

func getPorts(hostId string) *Ports {
	/* inputData := `/ethdev/list`
	cmd := exec.Command("python3", "/home/dsw/codes/p4-controller/dpdk-telemetry.py")
	cmd.Stdin = bytes.NewBufferString(inputData)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run() */

	host := hostsInfo[hostId]
	url := fmt.Sprintf("http://%s:%s/%s", host.IP, host.HTTPPort, "ports")
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	var ports Ports
	json.Unmarshal(data, &ports)
	return &ports
}

func getPortInfo(hostId string, portId int) *PortInfo {
	/* inputData := `/ethdev/info,` + strconv.Itoa(portId)
	cmd := exec.Command("python3", "/home/dsw/codes/p4-controller/dpdk-telemetry.py")
	cmd.Stdin = bytes.NewBufferString(inputData)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run() */

	host := hostsInfo[hostId]
	url := fmt.Sprintf("http://%s:%s/%s?portId=%v", host.IP, host.HTTPPort, "port_info", portId)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	var portInfo PortInfo
	var portInfoRoot PortInfoRoot

	json.Unmarshal(data, &portInfoRoot)
	portInfo.PortId = portId
	portInfo.MacAddr = portInfoRoot.MacAddr
	portInfo.Mtu = portInfoRoot.Mtu

	return &portInfo
}

func getPortStatus(hostId string, portId int) *PortStatus {
	/* inputData := `/ethdev/link_status,` + strconv.Itoa(portId)
	cmd := exec.Command("python3", "/home/dsw/codes/p4-controller/dpdk-telemetry.py")
	cmd.Stdin = bytes.NewBufferString(inputData)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run() */

	host := hostsInfo[hostId]
	url := fmt.Sprintf("http://%s:%s/%s?portId=%v", host.IP, host.HTTPPort, "port_status", portId)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	var portStatus PortStatus
	var portStatusRoot PortStatusRoot

	json.Unmarshal(data, &portStatusRoot)
	portStatus.PortId = portId
	portStatus.Status = portStatusRoot.Status
	portStatus.Speed = portStatusRoot.Speed
	portStatus.Duplex = portStatusRoot.Duplex

	return &portStatus
}

func getPortStats(hostId string, portId int) *PortStats {
	/* inputData := `/ethdev/stats,` + strconv.Itoa(portId)
	cmd := exec.Command("python3", "/home/dsw/codes/p4-controller/dpdk-telemetry.py")
	cmd.Stdin = bytes.NewBufferString(inputData)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Run() */

	host := hostsInfo[hostId]
	url := fmt.Sprintf("http://%s:%s/%s?portId=%v", host.IP, host.HTTPPort, "port_stats", portId)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println(err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
	}

	var portStats PortStats
	var portStatsRoot PortStatsRoot

	json.Unmarshal(data, &portStatsRoot)
	portStats.PortId = portId
	portStats.RxPackets = portStatsRoot.RxPackets
	portStats.TxPackets = portStatsRoot.TxPackets
	portStats.RxBytes = portStatsRoot.RxBytes
	portStats.TxBytes = portStatsRoot.TxBytes

	return &portStats
}
