package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/antoninbas/p4runtime-go-client/pkg/client"
	"github.com/gin-gonic/gin"
	p4_v1 "github.com/p4lang/p4runtime/go/p4/v1"
)

func setPipeconfHandler(ctx *gin.Context) {
	if notPrimary(ctx) {
		return
	}

	bin, _ := ctx.FormFile("bin")
	p4info, _ := ctx.FormFile("p4info")
	binPath := tmpDir + bin.Filename
	p4infoPath := tmpDir + p4info.Filename
	ctx.SaveUploadedFile(bin, binPath)
	ctx.SaveUploadedFile(p4info, p4infoPath)
	log.Println("saved bin and p4info.")

	if _, err := p4rt_ctl.SetFwdPipe(context.Background(), binPath, p4infoPath, 0); err != nil {
		msg := "setting pipeline config failed."
		ctx.JSON(http.StatusOK, gin.H{
			"msg": msg,
		})
		log.Println(msg)
	} else {
		msg := "setting pipeline config successfully."
		ctx.JSON(http.StatusOK, gin.H{
			"msg": msg,
		})
		log.Println(msg)
	}
}

func insertTableEntryExactHandler(ctx *gin.Context) {
	if notPrimary(ctx) {
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
		ctx.JSON(503, gin.H{
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
