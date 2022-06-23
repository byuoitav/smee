package commandcli

import (
	"fmt"
	"net/http"
	"strings"

	avcli "github.com/byuoitav/smee/proto"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func (c *Client) RemoveDevice(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		c.log.Warn("no room/device id provided; cancelling removal...")
		ctx.JSON(http.StatusBadRequest, "no room/device id")
		return
	}

	cookie := ctx.Request.Header.Get("Cookie")
	segments := strings.Split(cookie, ".")
	if len(segments) < 3 {
		c.log.Warn("no av-user specified; cancelling removal...")
		ctx.JSON(http.StatusBadRequest, "no av-user specified")
		return
	}

	netid, err := getUserFromJWT(segments[1])
	if err != nil || len(netid) == 0 {
		c.log.Warn("no av-user specified; cancelling removal...")
		ctx.JSON(http.StatusBadRequest, "no av-user specified")
		return
	}

	c.log.Debug(fmt.Sprintf("removing device with id: %s from monitoring", id))

	args := &avcli.ID{
		Id:          id,
		Designation: "prd",
	}

	auth := auth{
		token: c.cliToken,
		user:  netid,
	}

	_, err = c.cli.RemoveDeviceFromMonitoring(ctx.Request.Context(), args, grpc.PerRPCCredentials(auth))
	if err != nil {
		c.log.Warn(fmt.Sprintf("unable to remove device: %s from monitoring", id), zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, fmt.Sprintf("unable to remove device: %s from monitoring", id))
		return
	}

	ctx.JSON(http.StatusOK, fmt.Sprintf("removed device: %s", id))
}
