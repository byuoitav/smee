package commandcli

import (
	"fmt"
	"net/http"

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
	token, err := parseForCookie("smee", cookie)
	if err != nil {
		c.log.Warn("authorization not found; cancelling closure...")
		ctx.JSON(http.StatusBadRequest, "authorization not found")
		return
	}

	netid, err := getUserFromJWT(token)
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

	var results response
	_, err = c.cli.CloseMonitoringIssue(ctx.Request.Context(), args, grpc.PerRPCCredentials(auth))
	if err != nil {
		c.log.Warn("unable to close issue", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, fmt.Sprintf("unable to close issue for room: %s", id))
		return
	}

	results.successful(id)

	ctx.JSON(http.StatusOK, results.report())
}
