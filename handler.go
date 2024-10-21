package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/antoninbas/p4runtime-go-client/pkg/client"
	"github.com/gin-gonic/gin"
)

func setPipeconf(ctx *gin.Context) {
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

func insertTableEntryExact(ctx *gin.Context) {
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
