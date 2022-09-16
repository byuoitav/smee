package commandcli

import (
	"fmt"
	"io"
	"net/http"

	avcli "github.com/byuoitav/smee/proto"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func (c *Client) Swab(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		c.log.Warn("no room/device id provided; cancelling swab...")
		ctx.JSON(http.StatusBadRequest, "no room/device id")
		return
	}

	cookie := ctx.Request.Header.Get("Cookie")
	token, err := parseForCookie("smee", cookie)
	if err != nil {
		c.log.Warn("authorization not found; cancelling swab...")
		ctx.JSON(http.StatusBadRequest, "authorization not found")
		return
	}

	netid, err := getUserFromJWT(token)
	if err != nil || len(netid) == 0 {
		c.log.Warn("no av-user specified; cancelling swab...")
		ctx.JSON(http.StatusBadRequest, "no av-user specified")
		return
	}

	c.log.Debug(fmt.Sprintf("swabbing device/room with id: %s", id))

	// create args
	args := &avcli.ID{
		Id:          id,
		Designation: "prd",
	}

	auth := auth{
		token: c.cliToken,
		user:  netid,
	}

	stream, err := c.cli.Swab(ctx.Request.Context(), args, grpc.PerRPCCredentials(auth))
	if err != nil {
		c.log.Warn(fmt.Sprintf("unable to swab device/room: %s", id), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, fmt.Sprintf("unable to swab: %s", id))
		return
	}

	// recv on stream
	var results response

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			c.log.Warn("error receiving on the stream", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, fmt.Sprintf("failed to swab: %s", id))
			return
		}

		if resp.GetError() != "" {
			results.failed(resp.GetId())
		} else {
			results.successful(resp.GetId())
		}
	}

	ctx.JSON(http.StatusOK, results.report())
}
