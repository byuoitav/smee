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

func (c *Client) Swab(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		c.log.Warn("no room/device id provided; cancelling swab...")
		ctx.JSON(http.StatusBadRequest, "no room/device id")
		return
	}

	cookie := ctx.Request.Header.Get("Cookie")
	segments := strings.Split(cookie, ".")
	if len(segments) < 3 {
		c.log.Warn("no av-user specified; cancelling swab...")
		ctx.JSON(http.StatusBadRequest, "no av-user specified")
		return
	}

	netid, err := getUserFromJWT(segments[1])
	if err != nil || len(netid) == 0 {
		c.log.Warn("no av-user specified; cancelling swab...")
		ctx.JSON(http.StatusBadRequest, "no av-user specified")
		return
	}

	c.log.Debug(fmt.Sprintf("swabing device/room with id: %s", id))

	// create args
	args := &avcli.ID{
		Id:          id,
		Designation: "prd",
	}

	auth := auth{
		token: c.cliToken,
		user:  netid,
	}

	stream, err := c.cli.Swab(ctx.Request.Context(), args, grpc.PerRPCCredentials(auth)) // Todo: add auth
	if err != nil {
		c.log.Warn(fmt.Sprintf("unable to swab device/room: %s", id), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, "unable to swab")
		return
	}

	// recv on stream
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		} else if err != nil {
			c.log.Warn("error receiving on the stream", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, "error occurred while swabing")
			return
		}

		if resp.GetError() != "" {
			c.log.Warn("error 2 receiving on the stream", zap.Error(err))
			ctx.JSON(http.StatusInternalServerError, "error occurred while swabing")
			return
		}
	}

	ctx.JSON(http.StatusOK, fmt.Sprintf("much swab: %s", id))
}
