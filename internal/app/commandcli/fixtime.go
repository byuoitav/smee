package commandcli

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	avcli "github.com/byuoitav/smee/proto"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func (c *Client) FixTime(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		c.log.Warn("no room/device id provided; cancelling time fix...")
		ctx.JSON(http.StatusBadRequest, "no room/device id")
		return
	}

	cookie := ctx.Request.Header.Get("Cookie")
	token := strings.TrimPrefix(cookie, "smee=")
	netid, err := getUserFromJWT(token)
	if err != nil || len(netid) == 0 {
		c.log.Warn("no av-user specified; cancelling time fix...")
		ctx.JSON(http.StatusBadRequest, "no av-user specified")
		return
	}

	c.log.Debug(fmt.Sprintf("Fixing time for room/device: %s", id))

	args := &avcli.ID{
		Id:          id,
		Designation: "prd",
	}

	auth := auth{
		token: c.cliToken,
		user:  netid,
	}

	stream, err := c.cli.FixTime(ctx.Request.Context(), args, grpc.PerRPCCredentials(auth))
	if err != nil {
		c.log.Warn(fmt.Sprintf("unable to fix time for room/device: %s", id), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, "unable to fix time")
		return
	}

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			c.log.Warn("unable to recv from stream while syncing time", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, "error occurred while syncing time")
			return
		}

		if resp.GetError() != "" {
			c.log.Warn("unable to recv from stream while syncing time", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, "error occurred while syncing time")
			return
		}
	}

	ctx.JSON(http.StatusOK, "fixed time")
}
