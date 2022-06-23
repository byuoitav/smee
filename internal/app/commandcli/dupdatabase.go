package commandcli

import (
	"net/http"
	"strings"

	avcli "github.com/byuoitav/smee/proto"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func (c *Client) DuplicateDatabase(ctx *gin.Context) {
	src := ctx.Param("src")
	dest := ctx.Param("dest")
	if src == "" || dest == "" {
		c.log.Warn("missing room id; cancelling duplication...")
		ctx.JSON(http.StatusBadRequest, "missing room id")
		return
	}

	cookie := ctx.Request.Header.Get("Cookie")
	segments := strings.Split(cookie, ".")
	if len(segments) < 3 {
		c.log.Warn("no av-user specified; cancelling duplication...")
		ctx.JSON(http.StatusBadRequest, "no av-user specified")
		return
	}

	netid, err := getUserFromJWT(segments[1])
	if err != nil || len(netid) == 0 {
		c.log.Warn("no av-user specified; cancelling duplication...")
		ctx.JSON(http.StatusBadRequest, "no av-user specified")
		return
	}

	args := &avcli.CopyRoomRequest{
		Src:            src,
		Dst:            dest,
		SrcDesignation: "prd",
		DstDesignation: "prd",
	}

	auth := auth{
		token: c.cliToken,
		user:  netid,
	}

	_, err = c.cli.CopyRoom(ctx.Request.Context(), args, grpc.PerRPCCredentials(auth))
	if err != nil {
		c.log.Warn("unable to duplicate room", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, "unable to duplicate room")
		return
	}

	ctx.JSON(http.StatusOK, "duplication successful")
}
