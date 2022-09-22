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

func (c *Client) Float(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		c.log.Warn("no room/device id provided; cancelling float...")
		ctx.JSON(http.StatusBadRequest, "no room/device id")
		return
	}

	cookie := ctx.Request.Header.Get("Cookie")
	token, err := parseForCookie("smee", cookie)
	if err != nil {
		c.log.Warn("authorization not found; cancelling float...")
		ctx.JSON(http.StatusBadRequest, "authorization not found")
		return
	}

	netid, err := getUserFromJWT(token)
	if err != nil || len(netid) == 0 {
		c.log.Warn("no av-user specified; cancelling float...")
		ctx.JSON(http.StatusBadRequest, "no av-user specified")
		return
	}

	c.log.Debug(fmt.Sprintf("floating device/room with id: %s", id))

	// create args
	args := &avcli.ID{
		Id:          id,
		Designation: "prd",
	}

	auth := auth{
		token: c.cliToken,
		user:  netid,
	}

	// call float on cli
	stream, err := c.cli.Float(ctx.Request.Context(), args, grpc.PerRPCCredentials(auth))
	if err != nil {
		c.log.Warn(fmt.Sprintf("unable to float device/room: %s", id), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, "unable to float")
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
			ctx.JSON(http.StatusInternalServerError, "error occurred while floating")
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
