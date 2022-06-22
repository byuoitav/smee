package commandcli

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	avcli "github.com/byuoitav/av-cli"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func (c *Client) Sink(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		c.log.Warn("no room/device id provided; cancelling sink...")
		ctx.JSON(http.StatusBadRequest, "no room/device id")
		return
	}

	cookie := ctx.Request.Header.Get("Cookie")
	segments := strings.Split(cookie, ".")
	if len(segments) < 3 {
		c.log.Warn("no av-user specified; cancelling sink...")
		ctx.JSON(http.StatusBadRequest, "no av-user specified")
		return
	}

	netid, err := getUserFromJWT(segments[1])
	if err != nil || len(netid) == 0 {
		c.log.Warn("no av-user specified; cancelling sink...")
		ctx.JSON(http.StatusBadRequest, "no av-user specified")
		return
	}

	c.log.Debug(fmt.Sprintf("sinking device/room with id: %s", id))

	// create args
	args := &avcli.ID{
		Id:          id,
		Designation: "prd",
	}

	auth := auth{
		token: c.cliToken,
		user:  netid,
	}

	stream, err := c.cli.Sink(ctx.Request.Context(), args, grpc.PerRPCCredentials(auth)) // Todo: add auth
	if err != nil {
		c.log.Warn(fmt.Sprintf("unable to sink device/room: %s", id), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, "unable to sink")
		return
	}

	// recv on stream
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			ctx.JSON(http.StatusInternalServerError, "error occurred while sinking")
			return
		}

		if resp.GetError() != "" {
			ctx.JSON(http.StatusInternalServerError, "error occurred while sinking")
			return
		}
	}

	ctx.JSON(http.StatusOK, fmt.Sprintf("very sink: %s", id))
}
