package main

import (
	"context"
	"log"
	"net/http"

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
		ctx.IndentedJSON(http.StatusOK, gin.H{
			"msg": msg,
		})
		log.Println(msg)
	} else {
		msg := "setting pipeline config successfully."
		ctx.IndentedJSON(http.StatusOK, gin.H{
			"msg": msg,
		})
		log.Println(msg)
	}
}


