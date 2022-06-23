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

func (c *Client) CloseIssueByRoom(ctx *gin.Context) {
	id := ctx.Param("id")
	if id == "" {
		c.log.Warn("no room id provided; cancelling closure...")
		ctx.JSON(http.StatusBadRequest, "no room id")
		return
	}

	cookie := ctx.Request.Header.Get("Cookie")
	segments := strings.Split(cookie, ".")
	if len(segments) < 3 {
		c.log.Warn("no av-user specified; cancelling closure...")
		ctx.JSON(http.StatusBadRequest, "no av-user specified")
		return
	}

	netid, err := getUserFromJWT(segments[1])
	if err != nil || len(netid) == 0 {
		c.log.Warn("no av-user specified; cancelling closure...")
		ctx.JSON(http.StatusBadRequest, "no av-user specified")
		return
	}

	c.log.Debug(fmt.Sprintf("closing issue for room: %s", id))

	args := &avcli.ID{
		Id:          id,
		Designation: "prd",
	}

	auth := auth{
		token: c.cliToken,
		user:  netid,
	}

	_, err = c.cli.CloseMonitoringIssue(ctx.Request.Context(), args, grpc.PerRPCCredentials(auth))
	if err != nil {
		c.log.Warn("unable to close issue", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, fmt.Sprintf("unable to close issue for room: %s", id))
		return
	}

	ctx.JSON(http.StatusOK, fmt.Sprintf("closed issue for room: %s", id))
}
