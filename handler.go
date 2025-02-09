package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/antoninbas/p4runtime-go-client/pkg/client"
	"github.com/gin-gonic/gin"
	p4_v1 "github.com/p4lang/p4runtime/go/p4/v1"
	"github.com/redis/go-redis/v9"
)

const (
	zsetName = `filesHash`
)

func filesListHandler(ctx *gin.Context) {
	filesHash := redisClient.ZRevRange(context.Background(), zsetName, 0, -1).Val()
	files := make([]FileInfoString, 0, len(filesHash))
	for _, hash := range filesHash {
		fileInfo := FileInfo{}
		redisClient.HGetAll(context.Background(), hash).Scan(&fileInfo)
		t := time.Unix(fileInfo.Timestamp, 0)
		files = append(files,
			FileInfoString{fileInfo.FileName, fmt.Sprintf("%d", fileInfo.Size), t.Format("2006-01-02 15:04:05"), hash})
	}
	ctx.JSON(http.StatusOK, files)
}

func fileDeleteHandler(ctx *gin.Context) {
	hash := ctx.Param("hash")
	fileName, err := redisClient.HGet(context.Background(), hash, "fileName").Result()
	if err == redis.Nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"msg": "file not exist.",
		})
		return
	}

	filePath := tmpDir + fileName
	err = os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		log.Println(err.Error())
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": "delete file failed: " + err.Error(),
		})
		return
	}

	redisClient.Del(context.Background(), hash)
	redisClient.ZRem(context.Background(), zsetName, hash)
	ctx.JSON(http.StatusOK, gin.H{
		"msg": "delete file successfully.",
	})
}

func fileDownloadHandler(ctx *gin.Context) {
	hash := ctx.Param("hash")
	fileName, err := redisClient.HGet(context.Background(), hash, "fileName").Result()
	if err == redis.Nil {
		ctx.JSON(http.StatusNotFound, gin.H{
			"msg": "file not exist.",
		})
		return
	}

	filePath := tmpDir + fileName
	ctx.Header("Content-Disposition", "attachment; filename="+fileName)
	ctx.Header("Content-Type", "application/octet-stream")
	ctx.File(filePath)
}

func setPipeconfHandler(ctx *gin.Context) {
	if notPrimary() {
		fmt.Println("not primary.")
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": "controller not primary.",
		})
		ctx.Abort()
		return
	}

	bin, _ := ctx.FormFile("bin")
	p4info, _ := ctx.FormFile("p4info")

	binBytes := make([]byte, bin.Size)
	binFile, _ := bin.Open()
	count, _ := binFile.Read(binBytes)
	log.Println(count)
	log.Println(binBytes)

	p4infoBytes := make([]byte, p4info.Size)
	p4infoFile, _ := p4info.Open()
	count, _ = p4infoFile.Read(p4infoBytes)
	log.Println(count)
	log.Println(p4infoBytes)

	for i := 0; i < 3; i++ {
		if _, err := p4rt_ctl.SetFwdPipeFromBytes(context.Background(), binBytes, p4infoBytes, 0); err != nil {
			// restart infrap4d and reconnect.
			path := "/root/p4-example/setup_ports.sh"
			cmd := exec.Command("bash", path)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			var i int
			for i = 0; i < 3; i++ {
				if err := cmd.Run(); err == nil {
					break
				} else {
					time.Sleep(100 * time.Millisecond)
				}
			}
			if i == 3 {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"msg": "setting pipeline config failed. please check device status.",
				})
				return
			}
			time.Sleep(800 * time.Millisecond)
		} else {
			msg := "setting pipeline config successfully."
			ctx.JSON(http.StatusOK, gin.H{
				"msg": msg,
			})
			log.Println(msg)

			if redisClient != nil && redisClient.Ping(context.Background()).Err() == nil {
				binHash := getSha256String(binBytes)
				p4infoHash := getSha256String(p4infoBytes)
				now := time.Now().Unix()

				redisClient.ZAdd(context.Background(), zsetName,
					redis.Z{Score: float64(now), Member: binHash},
					redis.Z{Score: float64(now), Member: p4infoHash},
				)

				redisClient.HSet(context.Background(), binHash, FileInfo{bin.Filename, bin.Size, now})
				redisClient.HSet(context.Background(), p4infoHash, FileInfo{p4info.Filename, p4info.Size, now})

				binPath := tmpDir + bin.Filename
				p4infoPath := tmpDir + p4info.Filename
				os.WriteFile(binPath, binBytes, 0644)
				os.WriteFile(p4infoPath, p4infoBytes, 0644)

				log.Println("saved bin and p4info to Redis and disk.")
			}
			return
		}
	}

	ctx.JSON(http.StatusInternalServerError, gin.H{
		"msg": "setting pipeline config failed. please check device status.",
	})
}

func getSha256String(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func insertTableEntryExactHandler(ctx *gin.Context) {
	if notPrimary() {
		return
	}

	var tabelEntry TableEntryExact
	if err := ctx.ShouldBindJSON(&tabelEntry); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": err.Error(),
		})
		return
	}

	params := make([][]byte, 0, len(tabelEntry.Params))
	for _, param := range tabelEntry.Params {
		params = append(params, stringToByteSlice(param))
	}

	action := p4rt_ctl.NewTableActionDirect(tabelEntry.ActionName, params)

	mfs := make(map[string]client.MatchInterface)
	for k, v := range tabelEntry.MatchField {
		mfs[k] = &client.ExactMatch{Value: stringToByteSlice(v)}
	}

	entry := p4rt_ctl.NewTableEntry(tabelEntry.TableName, mfs, action, nil)
	if err := p4rt_ctl.InsertTableEntry(context.Background(), entry); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"msg": err.Error(),
		})
		return
	}

	fmt.Printf("tabelEntry: %v\n", tabelEntry)
	ctx.JSON(http.StatusOK, gin.H{
		"msg": "insert a table entry successfully.",
	})
}

func sendPacketOutHandler(ctx *gin.Context) {
	value := make([]byte, 1)
	value[0] = byte(1)
	packetOut := &p4_v1.PacketOut{Metadata: []*p4_v1.PacketMetadata{{MetadataId: 1, Value: value}}}
	err := p4rt_ctl.SendPacketOut(context.Background(), packetOut)
	if err != nil {
		ctx.JSON(503, err)
	} else {
		ctx.JSON(200, gin.H{
			"msg": "send packet out ok.",
		})
	}
}

func getTableEntriesByNameHandler(ctx *gin.Context) {
	tableName := ctx.Query("name")
	if tableName == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"msg": "table name is empty.",
		})
		return
	}

	if entries, err := p4rt_ctl.ReadTableEntryWildcard(context.Background(), tableName); err != nil {
		ctx.JSON(http.StatusServiceUnavailable, gin.H{
			"msg": err.Error(),
		})
	} else {
		resEntries := make([]TableEntryExact, 0)
		for _, entry := range entries {
			var resEntry TableEntryExact
			resEntry.TableName = tableName

			action := entry.GetAction().GetAction()
			pipeconf, _ := p4rt_ctl.GetFwdPipe(context.Background(), client.GetFwdPipeP4InfoAndCookie)
			resEntry.ActionName = getActionName(action, pipeconf)

			resEntry.Params = getActionParams(action)

			mfs := make(map[string]string)
			for _, fieldMatch := range entry.Match {
				mfs[getMatchFieldName(fieldMatch, tableName, pipeconf)] = byteSliceToString(fieldMatch.GetExact().Value)
			}
			resEntry.MatchField = mfs

			resEntries = append(resEntries, resEntry)
		}

		ctx.JSON(http.StatusOK, gin.H{
			"msg": resEntries,
		})
	}
}

func getActionName(action *p4_v1.Action, pipeconf *client.FwdPipeConfig) string {
	for _, ac := range pipeconf.P4Info.Actions {
		if ac.Preamble.Id == action.ActionId {
			return ac.Preamble.Name
		}
	}
	return ""
}

func getActionParams(action *p4_v1.Action) []string {
	res := make([]string, 0)
	for _, param := range action.Params {
		res = append(res, byteSliceToString(param.Value))
	}
	return res
}

func getMatchFieldName(field *p4_v1.FieldMatch, tableName string, pipeconf *client.FwdPipeConfig) string {
	for _, table := range pipeconf.P4Info.GetTables() {
		if table.Preamble.Name == tableName {
			for _, mf := range table.MatchFields {
				if field.FieldId == mf.Id {
					return mf.Name
				}
			}
		}
	}

	return ""
}
